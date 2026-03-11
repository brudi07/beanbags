package handlers

import (
	"database/sql"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateRound creates a new round for a match
// POST /api/rounds
func CreateRound(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			MatchID     int `json:"match_id" binding:"required"`
			RoundNumber int `json:"round_number" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec(`
			INSERT INTO rounds (match_id, round_number)
			VALUES (?, ?)
		`, req.MatchID, req.RoundNumber)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create round"})
			return
		}

		roundID, _ := result.LastInsertId()

		c.JSON(http.StatusCreated, gin.H{
			"id":           roundID,
			"match_id":     req.MatchID,
			"round_number": req.RoundNumber,
		})
	}
}

// AddThrow adds a throw to a round
// POST /api/rounds/:id/throws
func AddThrow(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roundID := c.Param("id")

		var throw models.Throw
		if err := c.ShouldBindJSON(&throw); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Calculate points based on result
		pointsMap := map[string]int{
			"hole":  3,
			"board": 1,
			"miss":  0,
			"ito":   0,
		}
		throw.PointsEarned = pointsMap[throw.Result]

		// Insert throw
		result, err := db.Exec(`
			INSERT INTO throws (
				round_id, player_id, team, throw_number, result,
				x_position, y_position, rotation, points_earned
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, roundID, throw.PlayerID, throw.Team, throw.ThrowNumber, throw.Result,
			throw.XPosition, throw.YPosition, throw.Rotation, throw.PointsEarned)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add throw"})
			return
		}

		throwID, _ := result.LastInsertId()

		c.JSON(http.StatusCreated, gin.H{
			"id":            throwID,
			"points_earned": throw.PointsEarned,
		})
	}
}

// GetRoundsForMatch retrieves all rounds and throws for a match
// Helper function used by GetMatch in games.go
func GetRoundsForMatch(db *sql.DB, matchID string) ([]models.Round, error) {
	rows, err := db.Query(`
		SELECT
			id, match_id, round_number, team1_score, team2_score,
			team1_points, team2_points, team1_busted, team2_busted, created_at
		FROM rounds
		WHERE match_id = ?
		ORDER BY round_number
	`, matchID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rounds := []models.Round{}
	for rows.Next() {
		var round models.Round
		if err := rows.Scan(
			&round.ID, &round.MatchID, &round.RoundNumber,
			&round.Team1Score, &round.Team2Score,
			&round.Team1Points, &round.Team2Points,
			&round.Team1Busted, &round.Team2Busted,
			&round.CreatedAt,
		); err != nil {
			continue
		}

		// Get throws for this round
		throwRows, err := db.Query(`
			SELECT
				id, round_id, player_id, team, throw_number,
				result, x_position, y_position, rotation, points_earned, created_at
			FROM throws
			WHERE round_id = ?
			ORDER BY throw_number
		`, round.ID)

		if err == nil {
			throws := []models.Throw{}
			for throwRows.Next() {
				var throw models.Throw
				if err := throwRows.Scan(
					&throw.ID, &throw.RoundID, &throw.PlayerID, &throw.Team,
					&throw.ThrowNumber, &throw.Result,
					&throw.XPosition, &throw.YPosition, &throw.Rotation,
					&throw.PointsEarned, &throw.CreatedAt,
				); err == nil {
					throws = append(throws, throw)
				}
			}
			round.Throws = throws
			throwRows.Close()
		}

		rounds = append(rounds, round)
	}

	return rounds, nil
}
