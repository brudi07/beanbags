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

// GetMyPlayer returns the current user's player profile with all team memberships
// GET /api/players/me
func GetMyPlayer(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		type TeamInfo struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			MemberCount int    `json:"member_count"`
		}
		type MyPlayer struct {
			ID     int        `json:"id"`
			UserID int        `json:"user_id"`
			Name   string     `json:"name"`
			Teams  []TeamInfo `json:"teams"`
		}

		var player MyPlayer
		err := db.QueryRow(`SELECT id, user_id, name FROM players WHERE user_id = ?`, userID).
			Scan(&player.ID, &player.UserID, &player.Name)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player profile not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player"})
			return
		}

		// Fetch all managed teams the player is on, with member count
		rows, err := db.Query(`
			SELECT t.id, t.name, COUNT(tm2.player_id) as member_count
			FROM team_members tm
			JOIN teams t ON tm.team_id = t.id
			LEFT JOIN team_members tm2 ON tm2.team_id = t.id
			WHERE tm.player_id = ? AND t.is_managed = 1
			GROUP BY t.id
		`, player.ID)
		if err == nil {
			defer rows.Close()
			player.Teams = []TeamInfo{}
			for rows.Next() {
				var ti TeamInfo
				if rows.Scan(&ti.ID, &ti.Name, &ti.MemberCount) == nil {
					player.Teams = append(player.Teams, ti)
				}
			}
		} else {
			player.Teams = []TeamInfo{}
		}

		c.JSON(http.StatusOK, player)
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

// GetMyStats returns lifetime throw stats computed from the database for the current user
// GET /api/players/me/stats
func GetMyStats(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var totalThrows, holes, boards, misses, itos, pointsContributed int
		err := db.QueryRow(`
			SELECT
				COUNT(*),
				SUM(CASE WHEN t.result = 'hole' THEN 1 ELSE 0 END),
				SUM(CASE WHEN t.result = 'board' THEN 1 ELSE 0 END),
				SUM(CASE WHEN t.result = 'miss' THEN 1 ELSE 0 END),
				SUM(CASE WHEN t.result = 'ito' THEN 1 ELSE 0 END),
				SUM(t.points_earned)
			FROM throws t
			JOIN players p ON t.player_id = p.id
			WHERE p.user_id = ?
		`, userID).Scan(&totalThrows, &holes, &boards, &misses, &itos, &pointsContributed)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"total_throws": 0, "holes": 0, "boards": 0,
				"misses": 0, "itos": 0, "points_contributed": 0, "accuracy": 0.0,
			})
			return
		}

		accuracy := 0.0
		if totalThrows > 0 {
			accuracy = float64(holes+boards) / float64(totalThrows) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"total_throws":       totalThrows,
			"holes":              holes,
			"boards":             boards,
			"misses":             misses,
			"itos":               itos,
			"points_contributed": pointsContributed,
			"accuracy":           accuracy,
		})
	}
}

// GetMyGames returns completed game history for the current user
// GET /api/players/me/games
func GetMyGames(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		rows, err := db.Query(`
			SELECT
				m.id, m.start_time,
				t1.name as team1_name,
				t2.name as team2_name,
				tw.name as winning_team_name,
				CASE WHEN EXISTS (
					SELECT 1 FROM players p WHERE p.team_id = m.team1_id AND p.user_id = ?
				) THEN 1 ELSE 2 END as user_team
			FROM matches m
			JOIN teams t1 ON m.team1_id = t1.id
			JOIN teams t2 ON m.team2_id = t2.id
			LEFT JOIN teams tw ON m.winning_team_id = tw.id
			WHERE m.status = 'completed'
			AND (
				EXISTS (SELECT 1 FROM players p WHERE p.team_id = m.team1_id AND p.user_id = ?)
				OR EXISTS (SELECT 1 FROM players p WHERE p.team_id = m.team2_id AND p.user_id = ?)
			)
			ORDER BY m.start_time DESC
			LIMIT 50
		`, userID, userID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game history"})
			return
		}
		defer rows.Close()

		type GameEntry struct {
			ID       int     `json:"id"`
			Date     string  `json:"date"`
			Team1    string  `json:"team1"`
			Team2    string  `json:"team2"`
			Winner   *string `json:"winner"`
			UserTeam int     `json:"user_team"`
		}

		games := []GameEntry{}
		for rows.Next() {
			var g GameEntry
			var winner sql.NullString
			if err := rows.Scan(&g.ID, &g.Date, &g.Team1, &g.Team2, &winner, &g.UserTeam); err == nil {
				if winner.Valid {
					g.Winner = &winner.String
				}
				games = append(games, g)
			}
		}

		c.JSON(http.StatusOK, games)
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
