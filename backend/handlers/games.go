package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateMatch creates a new match/game from player names
// POST /api/games
func CreateMatch(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req struct {
			Format  string `json:"format" binding:"required,oneof=1v1 2v2"`
			BestOf  int    `json:"bestOf"`
			Players struct {
				Team1 []string `json:"team1" binding:"required"`
				Team2 []string `json:"team2" binding:"required"`
			} `json:"players" binding:"required"`
			LeagueGameID *int `json:"leagueGameId"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get or create teams by name
		team1Name := strings.Join(req.Players.Team1, " & ")
		team2Name := strings.Join(req.Players.Team2, " & ")

		db.Exec(`INSERT OR IGNORE INTO teams (name) VALUES (?)`, team1Name)
		var team1ID int64
		db.QueryRow(`SELECT id FROM teams WHERE name = ?`, team1Name).Scan(&team1ID)

		db.Exec(`INSERT OR IGNORE INTO teams (name) VALUES (?)`, team2Name)
		var team2ID int64
		db.QueryRow(`SELECT id FROM teams WHERE name = ?`, team2Name).Scan(&team2ID)

		// Create player records for this game session
		for _, name := range req.Players.Team1 {
			db.Exec(`INSERT INTO players (user_id, name, team_id) VALUES (?, ?, ?)`, userID, name, team1ID)
		}
		for _, name := range req.Players.Team2 {
			db.Exec(`INSERT INTO players (user_id, name, team_id) VALUES (?, ?, ?)`, userID, name, team2ID)
		}

		// Look up league_id from league game if provided
		var leagueID *int
		if req.LeagueGameID != nil {
			var lid int
			if err := db.QueryRow(`SELECT league_id FROM league_games WHERE id = ?`, *req.LeagueGameID).Scan(&lid); err == nil {
				leagueID = &lid
			}
		}

		// Create the match
		result, err := db.Exec(`
			INSERT INTO matches (league_id, team1_id, team2_id, status)
			VALUES (?, ?, ?, 'active')
		`, leagueID, team1ID, team2ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create match"})
			return
		}

		matchID, _ := result.LastInsertId()

		// Link match to league game and mark it in progress
		if req.LeagueGameID != nil {
			db.Exec(`UPDATE league_games SET game_id = ?, status = 'in_progress' WHERE id = ?`, matchID, *req.LeagueGameID)
		}

		c.JSON(http.StatusCreated, gin.H{"id": matchID})
	}
}

// GetMatches returns a list of matches with optional filters
// GET /api/matches?league_id=X&tournament_id=Y
func GetMatches(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Query("league_id")
		tournamentID := c.Query("tournament_id")

		query := `
			SELECT
				id, league_id, tournament_id, team1_id, team2_id,
				winning_team_id, start_time, end_time, location, notes, status
			FROM matches
			WHERE 1=1
		`

		args := []interface{}{}
		if leagueID != "" {
			query += " AND league_id = ?"
			args = append(args, leagueID)
		}
		if tournamentID != "" {
			query += " AND tournament_id = ?"
			args = append(args, tournamentID)
		}

		query += " ORDER BY start_time DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch matches"})
			return
		}
		defer rows.Close()

		matches := []models.Match{}
		for rows.Next() {
			var m models.Match
			if err := rows.Scan(
				&m.ID, &m.LeagueID, &m.TournamentID,
				&m.Team1ID, &m.Team2ID, &m.WinningTeamID,
				&m.StartTime, &m.EndTime, &m.Location, &m.Notes, &m.Status,
			); err == nil {
				matches = append(matches, m)
			}
		}

		c.JSON(http.StatusOK, matches)
	}
}

// GetMatch retrieves a specific match with all rounds and throws
// GET /api/games/:id
func GetMatch(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		var match models.Match
		err := db.QueryRow(`
			SELECT
				id, league_id, tournament_id, team1_id, team2_id, winning_team_id,
				start_time, end_time, location, notes, status
			FROM matches
			WHERE id = ?
		`, matchID).Scan(
			&match.ID, &match.LeagueID, &match.TournamentID,
			&match.Team1ID, &match.Team2ID, &match.WinningTeamID,
			&match.StartTime, &match.EndTime, &match.Location, &match.Notes, &match.Status,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch match"})
			return
		}

		// Get rounds with throws using helper from rounds.go
		rounds, err := GetRoundsForMatch(db, matchID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rounds"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"match":  match,
			"rounds": rounds,
		})
	}
}

// CompleteGame marks a game as completed and saves final results
// POST /api/games/:id/complete
func CompleteGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		var req models.GameSubmitRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Update match
		_, err = tx.Exec(`
			UPDATE matches
			SET winning_team_id = ?, end_time = ?, status = 'completed'
			WHERE id = ?
		`, req.Winner, req.CompletedAt, matchID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
			return
		}

		// Save all rounds if not already saved
		for _, roundResult := range req.RoundHistory {
			// Check if round exists
			var roundID int
			err := tx.QueryRow(`
				SELECT id FROM rounds WHERE match_id = ? AND round_number = ?
			`, matchID, roundResult.Round).Scan(&roundID)

			if err == sql.ErrNoRows {
				// Create round
				result, err := tx.Exec(`
					INSERT INTO rounds (
						match_id, round_number, team1_score, team2_score,
						team1_points, team2_points, team1_busted, team2_busted
					) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
				`, matchID, roundResult.Round, 0, 0,
					roundResult.Team1Points, roundResult.Team2Points,
					roundResult.Team1Busted, roundResult.Team2Busted)

				if err != nil {
					continue
				}

				roundID64, _ := result.LastInsertId()
				roundID = int(roundID64)

				// Save throws
				for _, throw := range roundResult.Throws {
					_, _ = tx.Exec(`
						INSERT INTO throws (
							round_id, player_id, team, throw_number,
							result, x_position, y_position, rotation, points_earned
						) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
					`, roundID, throw.PlayerID, throw.Team, throw.ThrowNumber,
						throw.Result, throw.XPosition, throw.YPosition,
						throw.Rotation, throw.PointsEarned)
				}
			}
		}

		// Update league standings if this is a league match
		var leagueID *int
		var team1ID, team2ID int
		err = tx.QueryRow(`
			SELECT league_id, team1_id, team2_id FROM matches WHERE id = ?
		`, matchID).Scan(&leagueID, &team1ID, &team2ID)

		if err == nil && leagueID != nil {
			UpdateLeagueStandings(tx, *leagueID, team1ID, team2ID, req.FinalScore.Team1, req.FinalScore.Team2)
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save game"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game completed successfully"})
	}
}

// GetGameResults returns the results page data for a completed game
// GET /api/games/:id/results
func GetGameResults(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Reuse GetMatch logic - it returns all the data needed for results
		GetMatch(db)(c)
	}
}
