package handlers

import (
	"database/sql"
	"encoding/json"
)

// UpdateLeagueStandings updates team-based standings after a 2v2 match
func UpdateLeagueStandings(tx *sql.Tx, leagueID, team1ID, team2ID, team1Score, team2Score int) error {
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
		if err := updateTeamStats(tx, leagueID, team1ID, 0, 0, 1, team1Score, team2Score); err != nil {
			return err
		}
		if err := updateTeamStats(tx, leagueID, team2ID, 0, 0, 1, team2Score, team1Score); err != nil {
			return err
		}
	} else {
		if err := updateTeamStats(tx, leagueID, winner, 1, 0, 0, winnerScore, loserScore); err != nil {
			return err
		}
		if err := updateTeamStats(tx, leagueID, loser, 0, 1, 0, loserScore, winnerScore); err != nil {
			return err
		}
	}
	return nil
}

// updateTeamStats updates or inserts team statistics in league standings
func updateTeamStats(tx *sql.Tx, leagueID, teamID, win, loss, tie, scored, allowed int) error {
	diff := scored - allowed

	totalGames := win + loss + tie
	var winPct float64
	if totalGames > 0 {
		winPct = float64(win) / float64(totalGames) * 100
	}

	_, err := tx.Exec(`
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
	return err
}

// UpdatePlayerStandings updates player-based standings after a 1v1 match
func UpdatePlayerStandings(tx *sql.Tx, leagueID, player1ID, player2ID, score1, score2 int) error {
	var winner, loser int
	var winnerScore, loserScore int
	tie := score1 == score2

	if score1 > score2 {
		winner, loser = player1ID, player2ID
		winnerScore, loserScore = score1, score2
	} else if score2 > score1 {
		winner, loser = player2ID, player1ID
		winnerScore, loserScore = score2, score1
	}

	if tie {
		if err := updatePlayerStats(tx, leagueID, player1ID, 0, 0, 1, score1, score2); err != nil {
			return err
		}
		if err := updatePlayerStats(tx, leagueID, player2ID, 0, 0, 1, score2, score1); err != nil {
			return err
		}
	} else {
		if err := updatePlayerStats(tx, leagueID, winner, 1, 0, 0, winnerScore, loserScore); err != nil {
			return err
		}
		if err := updatePlayerStats(tx, leagueID, loser, 0, 1, 0, loserScore, winnerScore); err != nil {
			return err
		}
	}
	return nil
}

// updatePlayerStats updates or inserts player statistics in league standings (1v1)
func updatePlayerStats(tx *sql.Tx, leagueID, playerID, win, loss, tie, scored, allowed int) error {
	diff := scored - allowed

	totalGames := win + loss + tie
	var winPct float64
	if totalGames > 0 {
		winPct = float64(win) / float64(totalGames) * 100
	}

	// Try update first
	result, err := tx.Exec(`
		UPDATE league_standings
		SET wins = wins + ?, losses = losses + ?, ties = ties + ?,
		    points_for = points_for + ?, points_against = points_against + ?,
		    point_diff = point_diff + ?,
		    win_percentage = (wins + ?) * 100.0 / (wins + losses + ties + ? + ? + ?),
		    updated_at = CURRENT_TIMESTAMP
		WHERE league_id = ? AND player_id = ?
	`, win, loss, tie, scored, allowed, diff, win, win, loss, tie, leagueID, playerID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		return nil
	}

	// No existing row — insert
	_, err = tx.Exec(`
		INSERT INTO league_standings (
			league_id, player_id, wins, losses, ties,
			points_for, points_against, point_diff, win_percentage
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, leagueID, playerID, win, loss, tie, scored, allowed, diff, winPct)
	return err
}

// ParseLeagueGameFormat looks up the format and player IDs for a league game by match ID
func ParseLeagueGameFormat(db *sql.Tx, matchID string) (format string, team1PlayerIDs, team2PlayerIDs []int, leagueID int, err error) {
	var t1JSON, t2JSON string
	err = db.QueryRow(`
		SELECT l.format, lg.team1_player_ids, lg.team2_player_ids, lg.league_id
		FROM league_games lg
		JOIN leagues l ON lg.league_id = l.id
		WHERE lg.game_id = ?
	`, matchID).Scan(&format, &t1JSON, &t2JSON, &leagueID)
	if err != nil {
		return
	}
	json.Unmarshal([]byte(t1JSON), &team1PlayerIDs)
	json.Unmarshal([]byte(t2JSON), &team2PlayerIDs)
	return
}
