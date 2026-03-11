package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateLeague creates a new league
func CreateLeague(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req models.CreateLeagueRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec(`
			INSERT INTO leagues (
				name, description, organizer_id, format, games_per_match,
				max_teams, start_date, weeks_of_play, location, is_public
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, req.Name, req.Description, userID, req.Format, req.GamesPerMatch,
			req.MaxTeams, req.StartDate, req.WeeksOfPlay, req.Location, req.IsPublic)

		if err != nil {
			// Log the actual error for debugging
			println("CreateLeague error:", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create league: " + err.Error()})
			return
		}

		leagueID, _ := result.LastInsertId()

		c.JSON(http.StatusCreated, gin.H{
			"id":      leagueID,
			"message": "League created successfully",
		})
	}
}

// GetPublicLeagues returns all public leagues
func GetPublicLeagues(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT
				l.id, l.name, l.description, l.organizer_id, u.first_name || ' ' || u.last_name as organizer_name,
				l.format, l.games_per_match, l.max_teams, l.current_teams,
				l.start_date, l.weeks_of_play, l.location, l.is_public, l.status,
				l.created_at, l.updated_at
			FROM leagues l
			JOIN users u ON l.organizer_id = u.id
			WHERE l.is_public = 1
			ORDER BY l.created_at DESC
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leagues"})
			return
		}
		defer rows.Close()

		leagues := []models.League{}
		for rows.Next() {
			var league models.League
			if err := rows.Scan(
				&league.ID, &league.Name, &league.Description, &league.OrganizerID,
				&league.OrganizerName, &league.Format, &league.GamesPerMatch,
				&league.MaxTeams, &league.CurrentTeams, &league.StartDate,
				&league.WeeksOfPlay, &league.Location, &league.IsPublic,
				&league.Status, &league.CreatedAt, &league.UpdatedAt,
			); err != nil {
				continue
			}
			leagues = append(leagues, league)
		}

		c.JSON(http.StatusOK, leagues)
	}
}

// GetMyLeagues returns leagues user is part of
func GetMyLeagues(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Get player ID for this user
		var playerID int
		err := db.QueryRow(`SELECT id FROM players WHERE user_id = ?`, userID).Scan(&playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Player profile not found"})
			return
		}

		rows, err := db.Query(`
			SELECT DISTINCT
				l.id, l.name, l.description, l.organizer_id, u.first_name || ' ' || u.last_name as organizer_name,
				l.format, l.games_per_match, l.max_teams, l.current_teams,
				l.start_date, l.weeks_of_play, l.location, l.is_public, l.status,
				l.created_at, l.updated_at
			FROM leagues l
			JOIN users u ON l.organizer_id = u.id
			LEFT JOIN league_members lm ON l.id = lm.league_id
			WHERE l.organizer_id = ? OR lm.player_id = ?
			ORDER BY l.created_at DESC
		`, userID, playerID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leagues"})
			return
		}
		defer rows.Close()

		leagues := []models.League{}
		for rows.Next() {
			var league models.League
			if err := rows.Scan(
				&league.ID, &league.Name, &league.Description, &league.OrganizerID,
				&league.OrganizerName, &league.Format, &league.GamesPerMatch,
				&league.MaxTeams, &league.CurrentTeams, &league.StartDate,
				&league.WeeksOfPlay, &league.Location, &league.IsPublic,
				&league.Status, &league.CreatedAt, &league.UpdatedAt,
			); err != nil {
				continue
			}
			leagues = append(leagues, league)
		}

		c.JSON(http.StatusOK, leagues)
	}
}

// GetLeagueByID returns a specific league
func GetLeagueByID(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("id")

		var league models.League
		err := db.QueryRow(`
			SELECT
				l.id, l.name, l.description, l.organizer_id, u.first_name || ' ' || u.last_name as organizer_name,
				l.format, l.games_per_match, l.max_teams, l.current_teams,
				l.start_date, l.weeks_of_play, l.location, l.is_public, l.status,
				l.created_at, l.updated_at
			FROM leagues l
			JOIN users u ON l.organizer_id = u.id
			WHERE l.id = ?
		`, leagueID).Scan(
			&league.ID, &league.Name, &league.Description, &league.OrganizerID,
			&league.OrganizerName, &league.Format, &league.GamesPerMatch,
			&league.MaxTeams, &league.CurrentTeams, &league.StartDate,
			&league.WeeksOfPlay, &league.Location, &league.IsPublic,
			&league.Status, &league.CreatedAt, &league.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch league"})
			return
		}

		c.JSON(http.StatusOK, league)
	}
}

// JoinLeague adds a player (or team) to a league
func JoinLeague(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		leagueID := c.Param("id")

		// Get player and their team
		var playerID int
		var teamID sql.NullInt64
		err := db.QueryRow(`SELECT id, team_id FROM players WHERE user_id = ?`, userID).Scan(&playerID, &teamID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player profile not found"})
			return
		}

		// Get league info
		var currentTeams, maxTeams int
		var format string
		err = db.QueryRow(`SELECT current_teams, max_teams, format FROM leagues WHERE id = ?`, leagueID).
			Scan(&currentTeams, &maxTeams, &format)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		}

		if currentTeams >= maxTeams {
			c.JSON(http.StatusBadRequest, gin.H{"error": "League is full"})
			return
		}

		if format == "2v2" {
			if !teamID.Valid {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You need a team to join a 2v2 league. Go to Teams to create or join one."})
				return
			}

			// Get all players on the team
			rows, err := db.Query(`SELECT id FROM players WHERE team_id = ?`, teamID.Int64)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
				return
			}
			defer rows.Close()

			var teamPlayerIDs []int
			for rows.Next() {
				var pid int
				if rows.Scan(&pid) == nil {
					teamPlayerIDs = append(teamPlayerIDs, pid)
				}
			}

			if len(teamPlayerIDs) < 2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your team needs 2 players to join a 2v2 league"})
				return
			}

			// Add all team members to the league
			for _, pid := range teamPlayerIDs {
				_, err = db.Exec(`INSERT INTO league_members (league_id, player_id) VALUES (?, ?)`, leagueID, pid)
				if err != nil {
					c.JSON(http.StatusConflict, gin.H{"error": "Team is already in this league"})
					return
				}
			}
		} else {
			// 1v1: add single player
			_, err = db.Exec(`INSERT INTO league_members (league_id, player_id) VALUES (?, ?)`, leagueID, playerID)
			if err != nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Already a member or failed to join"})
				return
			}
		}

		db.Exec(`UPDATE leagues SET current_teams = current_teams + 1 WHERE id = ?`, leagueID)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully joined league"})
	}
}

// LeaveLeague removes a player (or team) from a league
func LeaveLeague(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		leagueID := c.Param("id")

		// Get player and their team
		var playerID int
		var teamID sql.NullInt64
		err := db.QueryRow(`SELECT id, team_id FROM players WHERE user_id = ?`, userID).Scan(&playerID, &teamID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player profile not found"})
			return
		}

		// Get league format
		var format string
		if err := db.QueryRow(`SELECT format FROM leagues WHERE id = ?`, leagueID).Scan(&format); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		}

		if format == "2v2" && teamID.Valid {
			// Remove all team members from the league
			rows, err := db.Query(`SELECT id FROM players WHERE team_id = ?`, teamID.Int64)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get team members"})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var pid int
				if rows.Scan(&pid) == nil {
					db.Exec(`DELETE FROM league_members WHERE league_id = ? AND player_id = ?`, leagueID, pid)
				}
			}
		} else {
			result, err := db.Exec(`DELETE FROM league_members WHERE league_id = ? AND player_id = ?`, leagueID, playerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave league"})
				return
			}
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Not a member of this league"})
				return
			}
		}

		db.Exec(`UPDATE leagues SET current_teams = current_teams - 1 WHERE id = ?`, leagueID)
		c.JSON(http.StatusOK, gin.H{"message": "Successfully left league"})
	}
}

// GetLeagueMembers returns all members of a league
func GetLeagueMembers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("id")

		rows, err := db.Query(`
			SELECT lm.league_id, lm.player_id, p.name as player_name, lm.joined_at
			FROM league_members lm
			JOIN players p ON lm.player_id = p.id
			WHERE lm.league_id = ?
			ORDER BY lm.joined_at
		`, leagueID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
			return
		}
		defer rows.Close()

		members := []models.LeagueMember{}
		for rows.Next() {
			var member models.LeagueMember
			if err := rows.Scan(
				&member.LeagueID, &member.PlayerID,
				&member.PlayerName, &member.JoinedAt,
			); err != nil {
				continue
			}
			members = append(members, member)
		}

		c.JSON(http.StatusOK, members)
	}
}

// GetLeagueSchedule returns scheduled games grouped by date
func GetLeagueSchedule(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("id")

		rows, err := db.Query(`
			SELECT
				id, league_id, match_number, game_number, scheduled_date,
				team1, team2, team1_player_ids, team2_player_ids,
				status, winning_team, game_id, created_at, updated_at
			FROM league_games
			WHERE league_id = ?
			ORDER BY scheduled_date, match_number, game_number
		`, leagueID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedule"})
			return
		}
		defer rows.Close()

		// Group games by date
		scheduleMap := make(map[string][]models.LeagueGame)

		for rows.Next() {
			var game models.LeagueGame
			var team1PlayerIDsJSON, team2PlayerIDsJSON string

			if err := rows.Scan(
				&game.ID, &game.LeagueID, &game.MatchNumber, &game.GameNumber,
				&game.ScheduledDate, &game.Team1, &game.Team2,
				&team1PlayerIDsJSON, &team2PlayerIDsJSON,
				&game.Status, &game.WinningTeam, &game.GameID,
				&game.CreatedAt, &game.UpdatedAt,
			); err != nil {
				continue
			}

			// Parse JSON arrays
			json.Unmarshal([]byte(team1PlayerIDsJSON), &game.Team1PlayerIDs)
			json.Unmarshal([]byte(team2PlayerIDsJSON), &game.Team2PlayerIDs)

			scheduleMap[game.ScheduledDate] = append(scheduleMap[game.ScheduledDate], game)
		}

		// Convert to array format
		schedule := []models.LeagueSchedule{}
		for date, games := range scheduleMap {
			schedule = append(schedule, models.LeagueSchedule{
				Date:  date,
				Games: games,
			})
		}

		// Return empty array if no games (not an error)
		if len(schedule) == 0 {
			c.JSON(http.StatusOK, []models.LeagueSchedule{})
			return
		}

		c.JSON(http.StatusOK, schedule)
	}
}

// GetLeagueStandings returns current standings
func GetLeagueStandings(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("id")

		rows, err := db.Query(`
			SELECT
				ls.league_id, ls.team_id, t.name as team_name,
				ls.wins, ls.losses, ls.ties, ls.points_for, ls.points_against,
				ls.point_diff, ls.win_percentage
			FROM league_standings ls
			JOIN teams t ON ls.team_id = t.id
			WHERE ls.league_id = ?
			ORDER BY ls.win_percentage DESC, ls.point_diff DESC
		`, leagueID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch standings"})
			return
		}
		defer rows.Close()

		standings := []models.LeagueStanding{}
		for rows.Next() {
			var standing models.LeagueStanding
			if err := rows.Scan(
				&standing.LeagueID, &standing.TeamID, &standing.TeamName,
				&standing.Wins, &standing.Losses, &standing.Ties,
				&standing.PointsFor, &standing.PointsAgainst,
				&standing.PointDiff, &standing.WinPercentage,
			); err != nil {
				continue
			}
			standings = append(standings, standing)
		}

		c.JSON(http.StatusOK, standings)
	}
}

// UpdateLeague allows organizer to update league details
// PATCH /api/leagues/:id
func UpdateLeague(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		leagueID := c.Param("id")

		// Check if user is organizer
		var organizerID int
		err := db.QueryRow(`SELECT organizer_id FROM leagues WHERE id = ?`, leagueID).Scan(&organizerID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		}

		if organizerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the organizer can update this league"})
			return
		}

		var req models.CreateLeagueRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(`
			UPDATE leagues
			SET name = ?, description = ?, format = ?, games_per_match = ?,
			    max_teams = ?, start_date = ?, weeks_of_play = ?,
			    location = ?, is_public = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, req.Name, req.Description, req.Format, req.GamesPerMatch,
			req.MaxTeams, req.StartDate, req.WeeksOfPlay,
			req.Location, req.IsPublic, leagueID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update league"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "League updated successfully"})
	}
}

// RescheduleLeagueGame allows organizer to change game date
// PATCH /api/leagues/:id/games/:gameId/reschedule
func RescheduleLeagueGame(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		leagueID := c.Param("id")
		gameID := c.Param("gameId")

		// Check if user is organizer
		var organizerID int
		err := db.QueryRow(`SELECT organizer_id FROM leagues WHERE id = ?`, leagueID).Scan(&organizerID)
		if err != nil || organizerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the organizer can reschedule games"})
			return
		}

		var req struct {
			ScheduledDate string `json:"scheduled_date" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = db.Exec(`
			UPDATE league_games
			SET scheduled_date = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND league_id = ?
		`, req.ScheduledDate, gameID, leagueID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reschedule game"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Game rescheduled successfully"})
	}
}
