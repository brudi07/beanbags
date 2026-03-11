package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateLeagueSchedule(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		leagueID := c.Param("id")

		var organizerID, gamesPerMatch, weeksOfPlay int
		var format, startDate string

		err := db.QueryRow(`
			SELECT organizer_id, format, games_per_match, weeks_of_play, start_date
			FROM leagues
			WHERE id = ?
		`, leagueID).Scan(&organizerID, &format, &gamesPerMatch, &weeksOfPlay, &startDate)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "League not found"})
			return
		}

		if organizerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only organizer can generate schedule"})
			return
		}

		var existing int
		db.QueryRow(`SELECT COUNT(*) FROM league_games WHERE league_id = ?`, leagueID).Scan(&existing)

		if existing > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule already exists"})
			return
		}

		if weeksOfPlay <= 0 {
			weeksOfPlay = 1
		}

		rows, err := db.Query(`
			SELECT lm.player_id, p.name
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

		type Member struct {
			PlayerID int
			Name     string
		}

		members := []Member{}

		for rows.Next() {
			var m Member
			rows.Scan(&m.PlayerID, &m.Name)
			members = append(members, m)
		}

		if len(members) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Need at least 2 members"})
			return
		}

		playersPerTeam := 1
		if format == "2v2" {
			playersPerTeam = 2
		}

		if len(members) < playersPerTeam*2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Need at least %d players for %s", playersPerTeam*2, format),
			})
			return
		}

		// -------------------------------
		// FIX 1: Parse date in LOCAL TZ
		// -------------------------------

		loc := time.Now().Location()

		startTime, err := time.ParseInLocation(
			"2006-01-02",
			startDate,
			loc,
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date"})
			return
		}

		startTime = time.Date(
			startTime.Year(),
			startTime.Month(),
			startTime.Day(),
			0, 0, 0, 0,
			loc,
		)

		// -------------------------------
		// Create Teams
		// -------------------------------

		teams := [][]Member{}

		for i := 0; i < len(members); i += playersPerTeam {
			if i+playersPerTeam <= len(members) {
				teams = append(teams, members[i:i+playersPerTeam])
			}
		}

		if len(teams) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough teams"})
			return
		}

		bye := false
		if len(teams)%2 == 1 {
			teams = append(teams, []Member{})
			bye = true
		}

		numTeams := len(teams)
		numRounds := numTeams - 1
		half := numTeams / 2

		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		matchNumber := 1

		// -------------------------------
		// FIX 2: Respect weeks_of_play
		// -------------------------------

		for week := 0; week < weeksOfPlay; week++ {

			round := week % numRounds

			gameDate := startTime.AddDate(0, 0, week*7).In(loc)
			gameDateStr := gameDate.Format("2006-01-02")

			for i := 0; i < half; i++ {

				team1 := teams[i]
				team2 := teams[numTeams-1-i]

				if bye && (len(team1) == 0 || len(team2) == 0) {
					continue
				}

				team1IDs := []int{}
				team2IDs := []int{}

				team1Name := ""
				team2Name := ""

				for idx, p := range team1 {
					team1IDs = append(team1IDs, p.PlayerID)
					if idx > 0 {
						team1Name += " & "
					}
					team1Name += p.Name
				}

				for idx, p := range team2 {
					team2IDs = append(team2IDs, p.PlayerID)
					if idx > 0 {
						team2Name += " & "
					}
					team2Name += p.Name
				}

				team1JSON, _ := json.Marshal(team1IDs)
				team2JSON, _ := json.Marshal(team2IDs)

				for gameNum := 1; gameNum <= gamesPerMatch; gameNum++ {

					_, err := tx.Exec(`
						INSERT INTO league_games (
							league_id,
							match_number,
							game_number,
							scheduled_date,
							team1,
							team2,
							team1_player_ids,
							team2_player_ids,
							status
						) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'scheduled')
					`,
						leagueID,
						matchNumber,
						gameNum,
						gameDateStr,
						team1Name,
						team2Name,
						string(team1JSON),
						string(team2JSON),
					)

					if err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
						return
					}

				}

				matchNumber++

			}

			// Rotate teams for next round
			last := teams[numTeams-1]
			copy(teams[2:], teams[1:numTeams-1])
			teams[1] = last

			// Reset rotation after full cycle
			if round == numRounds-1 {
				// optional: reset to original order if you want exact repeat cycles
			}

		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save schedule"})
			return
		}

		totalMatches := matchNumber - 1
		totalGames := totalMatches * gamesPerMatch

		c.JSON(http.StatusOK, gin.H{
			"message":       "Schedule generated successfully",
			"total_matches": totalMatches,
			"total_games":   totalGames,
		})

	}
}
