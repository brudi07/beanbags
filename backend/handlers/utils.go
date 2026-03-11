package handlers

import "database/sql"

// UpdateLeagueStandings updates standings after a match is completed
func UpdateLeagueStandings(tx *sql.Tx, leagueID, team1ID, team2ID, team1Score, team2Score int) {
	var winner, loser int
	var winnerScore, loserScore int
	tie := team1Score == team2Score

	if team1Score > team2Score {
		winner, loser = team1ID, team2ID
		winnerScore, loserScore = team1Score, team2Score
	} else if team2Score > team1Score {
		winner, loser = team2ID, team1ID
		winnerScore, loserScore = team2Score, team1Score
	}

	if tie {
		updateTeamStats(tx, leagueID, team1ID, 0, 0, 1, team1Score, team2Score)
		updateTeamStats(tx, leagueID, team2ID, 0, 0, 1, team2Score, team1Score)
	} else {
		updateTeamStats(tx, leagueID, winner, 1, 0, 0, winnerScore, loserScore)
		updateTeamStats(tx, leagueID, loser, 0, 1, 0, loserScore, winnerScore)
	}
}

// updateTeamStats updates or inserts team statistics in league standings
func updateTeamStats(tx *sql.Tx, leagueID, teamID, win, loss, tie, scored, allowed int) {
	diff := scored - allowed

	// Calculate win percentage
	totalGames := win + loss + tie
	var winPct float64
	if totalGames > 0 {
		winPct = float64(win) / float64(totalGames) * 100
	}

	tx.Exec(`
		INSERT INTO league_standings (
			league_id, team_id, wins, losses, ties,
			points_for, points_against, point_diff, win_percentage
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(league_id, team_id) DO UPDATE
		SET wins = wins + ?,
		    losses = losses + ?,
		    ties = ties + ?,
		    points_for = points_for + ?,
		    points_against = points_against + ?,
		    point_diff = point_diff + ?,
		    win_percentage = (wins + ?) * 100.0 / (wins + losses + ties + ? + ? + ?),
		    updated_at = CURRENT_TIMESTAMP;
	`, leagueID, teamID, win, loss, tie, scored, allowed, diff, winPct,
		win, loss, tie, scored, allowed, diff, win, win, loss, tie)
}
