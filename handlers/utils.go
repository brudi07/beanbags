// handlers/utils.go
package handlers

import (
	"database/sql"
	"math"
)

func updateLeagueStandings(db *sql.DB, leagueID, team1ID, team2ID, team1Score, team2Score int) {
	tx, _ := db.Begin()
	defer tx.Commit()

	var winner, loser int
	var tie bool

	if team1Score > team2Score {
		winner, loser = team1ID, team2ID
	} else if team2Score > team1Score {
		winner, loser = team2ID, team1ID
	} else {
		tie = true
	}

	if tie {
		updateTeamStats(tx, leagueID, team1ID, 0, 0, 1, team1Score, team2Score)
		updateTeamStats(tx, leagueID, team2ID, 0, 0, 1, team2Score, team1Score)
	} else {
		updateTeamStats(tx, leagueID, winner, 1, 0, 0, int(math.Max(float64(team1Score), float64(team2Score))), int(math.Min(float64(team1Score), float64(team2Score))))
		updateTeamStats(tx, leagueID, loser, 0, 1, 0, int(math.Min(float64(team1Score), float64(team2Score))), int(math.Max(float64(team1Score), float64(team2Score))))
	}
}

func updateTeamStats(tx *sql.Tx, leagueID, teamID, win, loss, tie, scored, allowed int) {
	diff := scored - allowed
	tx.Exec(`
		INSERT INTO league_standings (league_id, team_id, wins, losses, ties, points_for, points_against, point_diff)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(league_id, team_id) DO UPDATE
		SET wins = wins + ?,
		    losses = losses + ?,
		    ties = ties + ?,
		    points_for = points_for + ?,
		    points_against = points_against + ?,
		    point_diff = point_diff + ?,
		    updated_at = CURRENT_TIMESTAMP;
	`, leagueID, teamID, win, loss, tie, scored, allowed, diff,
		win, loss, tie, scored, allowed, diff)
}
