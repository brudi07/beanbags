package handlers

import (
	"database/sql"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateMatch creates a league match. Pickup games are handled client-side only.
// POST /api/games
func CreateMatch(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req struct {
			Format       string `json:"format" binding:"required,oneof=1v1 2v2"`
			BestOf       int    `json:"bestOf"`
			Players      struct {
				Team1 []string `json:"team1" binding:"required"`
				Team2 []string `json:"team2" binding:"required"`
			} `json:"players" binding:"required"`
			LeagueGameID *int `json:"leagueGameId"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.LeagueGameID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "League game ID required"})
			return
		}

		// Look up the league game to get league_id and team names
		var leagueID int
		var team1Name, team2Name string
		if err := db.QueryRow(`SELECT league_id, team1, team2 FROM league_games WHERE id = ?`, *req.LeagueGameID).
			Scan(&leagueID, &team1Name, &team2Name); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "League game not found"})
			return
		}

		// Look up managed team IDs by name
		var team1ID, team2ID sql.NullInt64
		db.QueryRow(`SELECT id FROM teams WHERE name = ? AND is_managed = 1`, team1Name).Scan(&team1ID)
		db.QueryRow(`SELECT id FROM teams WHERE name = ? AND is_managed = 1`, team2Name).Scan(&team2ID)

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
		if _, err := db.Exec(`UPDATE league_games SET game_id = ?, status = 'in_progress' WHERE id = ?`, matchID, *req.LeagueGameID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update league game status"})
			return
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

		// Fetch actual team IDs to resolve the winning team FK
		var leagueID sql.NullInt64
		var team1ID, team2ID sql.NullInt64
		if err = tx.QueryRow(`
			SELECT league_id, team1_id, team2_id FROM matches WHERE id = ?
		`, matchID).Scan(&leagueID, &team1ID, &team2ID); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
			return
		}

		var winningTeamID sql.NullInt64
		if req.Winner == 2 {
			winningTeamID = team2ID
		} else {
			winningTeamID = team1ID
		}

		// Update match
		_, err = tx.Exec(`
			UPDATE matches
			SET winning_team_id = ?, end_time = ?, status = 'completed'
			WHERE id = ?
		`, winningTeamID, req.CompletedAt, matchID)

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

		// Update league standings
		if leagueID.Valid {
			lid := int(leagueID.Int64)
			if team1ID.Valid && team2ID.Valid {
				// 2v2: standings by team
				if standingsErr := UpdateLeagueStandings(tx, lid, int(team1ID.Int64), int(team2ID.Int64), req.FinalScore.Team1, req.FinalScore.Team2); standingsErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update league standings"})
					return
				}
			} else {
				// 1v1: standings by player — look up player IDs from the league game
				_, t1Players, t2Players, _, lgErr := ParseLeagueGameFormat(tx, matchID)
				if lgErr == nil && len(t1Players) > 0 && len(t2Players) > 0 {
					if standingsErr := UpdatePlayerStandings(tx, lid, t1Players[0], t2Players[0], req.FinalScore.Team1, req.FinalScore.Team2); standingsErr != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update league standings"})
						return
					}
				}
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save game"})
			return
		}

		// Mark the league game as completed and store final scores (best-effort, outside transaction)
		db.Exec(`
			UPDATE league_games
			SET status = 'completed', winning_team = ?, team1_score = ?, team2_score = ?
			WHERE game_id = ?
		`, req.Winner, req.FinalScore.Team1, req.FinalScore.Team2, matchID)

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
