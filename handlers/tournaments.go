package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateTournament handles POST /tournaments
func CreateTournament(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t models.Tournament
		if err := c.ShouldBindJSON(&t); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec(`
			INSERT INTO tournaments (name, start_date, end_date, location, league_id, format)
			VALUES (?, ?, ?, ?, ?, ?)`,
			t.Name, t.StartDate, t.EndDate, t.Location, t.LeagueID, t.Format)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, _ := res.LastInsertId()
		t.ID = int(id)

		// Insert participating teams
		for _, teamID := range t.TeamIDs {
			_, err := tx.Exec(`
				INSERT INTO tournament_teams (tournament_id, team_id)
				VALUES (?, ?)`, t.ID, teamID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, t)
	}
}

// GetTournaments handles GET /tournaments
func GetTournaments(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, start_date, end_date, location, league_id, format FROM tournaments`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var tournaments []models.Tournament
		for rows.Next() {
			var t models.Tournament
			var leagueID sql.NullInt64
			if err := rows.Scan(&t.ID, &t.Name, &t.StartDate, &t.EndDate, &t.Location, &leagueID, &t.Format); err != nil {
				continue
			}
			if leagueID.Valid {
				val := int(leagueID.Int64)
				t.LeagueID = &val
			}

			// Fetch team IDs for this tournament
			teamRows, _ := db.Query(`SELECT team_id FROM tournament_teams WHERE tournament_id = ?`, t.ID)
			for teamRows.Next() {
				var teamID int
				teamRows.Scan(&teamID)
				t.TeamIDs = append(t.TeamIDs, teamID)
			}
			teamRows.Close()

			tournaments = append(tournaments, t)
		}

		c.JSON(http.StatusOK, tournaments)
	}
}

// GenerateBracket handles POST /tournaments/:id/bracket
func GenerateBracket(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
			return
		}

		var format string
		err = db.QueryRow(`SELECT format FROM tournaments WHERE id = ?`, tournamentID).Scan(&format)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tournament not found"})
			return
		}

		// Load all participating teams
		rows, err := db.Query(`SELECT team_id FROM tournament_teams WHERE tournament_id = ?`, tournamentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var teamIDs []int
		for rows.Next() {
			var id int
			rows.Scan(&id)
			teamIDs = append(teamIDs, id)
		}

		if err := generateBracket(db, tournamentID, teamIDs, format); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Bracket generated successfully"})
	}
}

// GetBracket handles GET /tournaments/:id/bracket
func GetBracket(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		rows, err := db.Query(`
			SELECT id, team1_id, team2_id, team1_score, team2_score
			FROM matches WHERE tournament_id = ? ORDER BY id`, tournamentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var matches []map[string]interface{}
		for rows.Next() {
			var id, team1, team2, s1, s2 sql.NullInt64
			rows.Scan(&id, &team1, &team2, &s1, &s2)

			matches = append(matches, gin.H{
				"id":          id.Int64,
				"team1_id":    team1.Int64,
				"team2_id":    team2.Int64,
				"team1_score": s1.Int64,
				"team2_score": s2.Int64,
			})
		}

		c.JSON(http.StatusOK, matches)
	}
}

func generateBracket(db *sql.DB, tournamentID int, teamIDs []int, format string) error {
	if len(teamIDs) < 2 {
		return fmt.Errorf("need at least two teams to generate a bracket")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Commit()

	// Add a bye if needed
	if len(teamIDs)%2 != 0 {
		teamIDs = append(teamIDs, 0)
	}

	// Round 1 - Winners Bracket
	for i := 0; i < len(teamIDs); i += 2 {
		team1 := teamIDs[i]
		team2 := teamIDs[i+1]

		if team2 == 0 {
			// Handle bye
			_, err := tx.Exec(`
				INSERT INTO matches (tournament_id, team1_id, team2_id, team1_score, team2_score, completed, round, bracket)
				VALUES (?, ?, NULL, 21, 0, 1, 1, 'winners')`,
				tournamentID, team1)
			if err != nil {
				tx.Rollback()
				return err
			}
			continue
		}

		_, err := tx.Exec(`
			INSERT INTO matches (tournament_id, team1_id, team2_id, team1_score, team2_score, completed, round, bracket)
			VALUES (?, ?, ?, 0, 0, 0, 1, 'winners')`,
			tournamentID, team1, team2)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if format == "double" {
		// Placeholder for losers bracket
		_, err := tx.Exec(`
			INSERT INTO matches (tournament_id, round, bracket, completed)
			VALUES (?, 1, 'losers', 0);
		`, tournamentID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func advanceTournamentRound(db *sql.DB, tournamentID int, format string) error {
	rows, err := db.Query(`
		SELECT id, team1_id, team2_id, team1_score, team2_score, round, bracket
		FROM matches
		WHERE tournament_id = ? AND completed = 1
	`, tournamentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type Result struct {
		ID      int
		Winner  int
		Loser   int
		Round   int
		Bracket string
	}
	var results []Result

	for rows.Next() {
		var m Result
		var team1, team2, score1, score2 int
		rows.Scan(&m.ID, &team1, &team2, &score1, &score2, &m.Round, &m.Bracket)
		if score1 >= score2 {
			m.Winner = team1
			m.Loser = team2
		} else {
			m.Winner = team2
			m.Loser = team1
		}
		results = append(results, m)
	}

	tx, _ := db.Begin()
	defer tx.Commit()

	nextRound := results[0].Round + 1

	// Create new matches in winners bracket
	var winners []int
	for _, r := range results {
		if r.Bracket == "winners" {
			winners = append(winners, r.Winner)
		}
	}
	for i := 0; i < len(winners); i += 2 {
		if i+1 >= len(winners) {
			break
		}
		tx.Exec(`
			INSERT INTO matches (tournament_id, team1_id, team2_id, round, bracket)
			VALUES (?, ?, ?, ?, 'winners')`, tournamentID, winners[i], winners[i+1], nextRound)
	}

	if format == "double" {
		// Losers drop to losers bracket
		var losers []int
		for _, r := range results {
			if r.Bracket == "winners" {
				losers = append(losers, r.Loser)
			}
		}
		for i := 0; i < len(losers); i += 2 {
			if i+1 >= len(losers) {
				break
			}
			tx.Exec(`
				INSERT INTO matches (tournament_id, team1_id, team2_id, round, bracket)
				VALUES (?, ?, ?, ?, 'losers')`, tournamentID, losers[i], losers[i+1], nextRound)
		}
	}

	return nil
}
