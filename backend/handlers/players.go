package handlers

import (
	"database/sql"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreatePlayer creates a new player profile
// POST /api/players
func CreatePlayer(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req struct {
			Name   string `json:"name" binding:"required"`
			TeamID *int   `json:"team_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec(`
			INSERT INTO players (user_id, name, team_id)
			VALUES (?, ?, ?)
		`, userID, req.Name, req.TeamID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
			return
		}

		playerID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": playerID, "name": req.Name})
	}
}

// GetPlayers returns all players
// GET /api/players
func GetPlayers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, user_id, name, team_id, created_at
			FROM players
			ORDER BY name
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch players"})
			return
		}
		defer rows.Close()

		players := []models.Player{}
		for rows.Next() {
			var player models.Player
			if err := rows.Scan(
				&player.ID, &player.UserID, &player.Name,
				&player.TeamID, &player.CreatedAt,
			); err == nil {
				players = append(players, player)
			}
		}

		c.JSON(http.StatusOK, players)
	}
}

// GetPlayer returns a specific player by ID
// GET /api/players/:id
func GetPlayer(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.Param("id")

		var player models.Player
		err := db.QueryRow(`
			SELECT id, user_id, name, team_id, created_at
			FROM players
			WHERE id = ?
		`, playerID).Scan(
			&player.ID, &player.UserID, &player.Name,
			&player.TeamID, &player.CreatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player"})
			return
		}

		c.JSON(http.StatusOK, player)
	}
}

// GetPlayerStats returns statistics for a player
// GET /api/players/:id/stats?league_id=X
func GetPlayerStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.Param("id")
		leagueID := c.Query("league_id") // Optional filter

		query := `
			SELECT
				id, player_id, league_id, total_throws, holes, boards,
				misses, itos, busts, points_contributed, accuracy,
				points_per_round, differential_per_round, updated_at
			FROM player_stats
			WHERE player_id = ?
		`

		args := []interface{}{playerID}
		if leagueID != "" {
			query += " AND league_id = ?"
			args = append(args, leagueID)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
			return
		}
		defer rows.Close()

		stats := []models.PlayerStats{}
		for rows.Next() {
			var stat models.PlayerStats
			if err := rows.Scan(
				&stat.ID, &stat.PlayerID, &stat.LeagueID, &stat.TotalThrows,
				&stat.Holes, &stat.Boards, &stat.Misses, &stat.ITOs,
				&stat.Busts, &stat.PointsContributed, &stat.Accuracy,
				&stat.PointsPerRound, &stat.DifferentialPerRound, &stat.UpdatedAt,
			); err == nil {
				stats = append(stats, stat)
			}
		}

		c.JSON(http.StatusOK, stats)
	}
}

// GetPlayerHeatmap returns throw position data for visualization
// GET /api/players/:id/heatmap?league_id=X
func GetPlayerHeatmap(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		playerID := c.Param("id")
		leagueID := c.Query("league_id") // Optional filter

		query := `
			SELECT t.result, t.x_position, t.y_position, t.points_earned, t.created_at
			FROM throws t
			JOIN rounds r ON t.round_id = r.id
			JOIN matches m ON r.match_id = m.id
			WHERE t.player_id = ?
		`

		args := []interface{}{playerID}
		if leagueID != "" {
			query += " AND m.league_id = ?"
			args = append(args, leagueID)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch heatmap data"})
			return
		}
		defer rows.Close()

		type HeatmapPoint struct {
			Result       string  `json:"result"`
			X            float64 `json:"x"`
			Y            float64 `json:"y"`
			PointsEarned int     `json:"points_earned"`
			Timestamp    string  `json:"timestamp"`
		}

		points := []HeatmapPoint{}
		for rows.Next() {
			var point HeatmapPoint
			if err := rows.Scan(
				&point.Result, &point.X, &point.Y,
				&point.PointsEarned, &point.Timestamp,
			); err == nil {
				points = append(points, point)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"player_id": playerID,
			"throws":    points,
			"count":     len(points),
		})
	}
}
