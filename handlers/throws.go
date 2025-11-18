package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ThrowInput struct {
	PlayerID    int     `json:"player_id" binding:"required"`
	ThrowNumber int     `json:"throw_number" binding:"required"`
	ThrowType   string  `json:"throw_type" binding:"required"`
	XPosition   float64 `json:"x_position"`
	YPosition   float64 `json:"y_position"`
}

// POST /rounds/:id/throws
func (h *Handler) AddThrow(c *gin.Context) {
	roundID := c.Param("id")
	var input ThrowInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine points based on throw type
	points := map[string]int{
		"in_hole":         3,
		"on_board":        1,
		"off_board":       0,
		"intentional_off": 0,
	}[input.ThrowType]

	// Check current total score to see if this causes a bust
	var currentTeamScore, matchID int
	err := h.DB.QueryRow(`
		SELECT r.match_id,
		       CASE
		           WHEN p.team_id = m.team1_id THEN r.team1_score
		           ELSE r.team2_score
		       END as team_score
		FROM rounds r
		JOIN matches m ON r.match_id = m.id
		JOIN players p ON p.id = ?
		WHERE r.id = ?`, input.PlayerID, roundID).Scan(&matchID, &currentTeamScore)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get current score: " + err.Error()})
		return
	}

	causedBust := false
	newScore := currentTeamScore + points
	if newScore > 21 {
		causedBust = true
		points = 0 // bust causes zero gain
	}

	// Insert the throw record
	res, err := h.DB.Exec(`
		INSERT INTO throws (round_id, player_id, throw_number, throw_type, x_position, y_position, points_earned, caused_bust)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		roundID, input.PlayerID, input.ThrowNumber, input.ThrowType, input.XPosition, input.YPosition, points, causedBust)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	throwID, _ := res.LastInsertId()

	// Update stats and round scores
	if causedBust {
		handleBust(h.DB, roundID, input.PlayerID)
	} else {
		updateScoresAndStats(h.DB, roundID, input.PlayerID, points, input.ThrowType)
	}

	c.JSON(http.StatusCreated, gin.H{
		"throw_id":    throwID,
		"points":      points,
		"caused_bust": causedBust,
	})
}

// --------------------
// Internal helpers
// --------------------

func updateScoresAndStats(db *sql.DB, roundID string, playerID int, points int, throwType string) {
	tx, _ := db.Begin()
	defer tx.Commit()

	// Update team score for the round
	_, _ = tx.Exec(`
		UPDATE rounds
		SET team1_score = CASE
		    WHEN (SELECT team_id FROM players WHERE id = ?) =
		         (SELECT team1_id FROM matches WHERE id = rounds.match_id)
		    THEN team1_score + ?
		    ELSE team1_score
		END,
		team2_score = CASE
		    WHEN (SELECT team_id FROM players WHERE id = ?) =
		         (SELECT team2_id FROM matches WHERE id = rounds.match_id)
		    THEN team2_score + ?
		    ELSE team2_score
		END
		WHERE id = ?;`,
		playerID, points, playerID, points, roundID)

	// Update player_stats table
	fieldMap := map[string]string{
		"in_hole":         "bags_in_hole",
		"on_board":        "bags_on_board",
		"off_board":       "bags_off_board",
		"intentional_off": "intentional_offs",
	}

	field := fieldMap[throwType]

	_, _ = tx.Exec(fmt.Sprintf(`
		INSERT INTO player_stats (player_id, total_throws, %s, points_total)
		VALUES (?, 1, 1, ?)
		ON CONFLICT(player_id) DO UPDATE
		SET total_throws = total_throws + 1,
		    %s = %s + 1,
		    points_total = points_total + ?,
		    average_points_per_round = points_total * 1.0 / total_throws,
		    updated_at = CURRENT_TIMESTAMP;
	`, field, field, field), playerID, points, playerID, points)
}

// When a bust occurs
func handleBust(db *sql.DB, roundID string, playerID int) {
	tx, _ := db.Begin()
	defer tx.Commit()

	// Mark the round as ended in bust
	_, _ = tx.Exec(`UPDATE rounds SET ended_in_bust = 1 WHERE id = ?`, roundID)

	// Increment bust stat for the player
	_, _ = tx.Exec(`
		INSERT INTO player_stats (player_id, total_throws, busts_caused)
		VALUES (?, 1, 1)
		ON CONFLICT(player_id) DO UPDATE
		SET total_throws = total_throws + 1,
		    busts_caused = busts_caused + 1,
		    updated_at = CURRENT_TIMESTAMP;
	`, playerID)

	// Auto-mark remaining throws for both teams as intentional offs
	_, _ = tx.Exec(`
		UPDATE throws
		SET auto_thrown_off = 1, throw_type = 'intentional_off'
		WHERE round_id = ? AND caused_bust = 0 AND id NOT IN (
		    SELECT id FROM throws WHERE round_id = ? AND caused_bust = 1
		);
	`, roundID, roundID)
}

// --------------------
// Analytics / Heatmap
// --------------------

// GET /players/:id/heatmap?start=YYYY-MM-DD&end=YYYY-MM-DD
func (h *Handler) GetPlayerHeatmap(c *gin.Context) {
	playerID := c.Param("id")
	start := c.Query("start")
	end := c.Query("end")

	if start == "" || end == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end date required"})
		return
	}

	rows, err := h.DB.Query(`
		SELECT x_position, y_position, throw_type, created_at
		FROM throws
		WHERE player_id = ?
		  AND DATE(created_at) BETWEEN DATE(?) AND DATE(?)
		  AND x_position IS NOT NULL
		  AND y_position IS NOT NULL;
	`, playerID, start, end)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type HeatPoint struct {
		X         float64   `json:"x"`
		Y         float64   `json:"y"`
		Type      string    `json:"type"`
		Timestamp time.Time `json:"timestamp"`
	}

	var points []HeatPoint
	for rows.Next() {
		var hp HeatPoint
		rows.Scan(&hp.X, &hp.Y, &hp.Type, &hp.Timestamp)
		points = append(points, hp)
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id": playerID,
		"start":     start,
		"end":       end,
		"throws":    points,
	})
}
