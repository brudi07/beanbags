package handlers

import (
	"brudi07/beanbags/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateMatch(c *gin.Context) {
	var m models.Match
	if err := c.BindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// allow both optional
	res, err := h.DB.Exec(`
		INSERT INTO matches (league_id, tournament_id, team1_id, team2_id, location, notes)
		VALUES (?, ?, ?, ?, ?, ?)`,
		m.LeagueID, m.TournamentID, m.Team1ID, m.Team2ID, m.Location, m.Notes,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	m.ID = int(id)
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetMatches(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, team1_id, team2_id, winning_team_id, start_time, end_time, location, notes FROM matches`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		var m models.Match
		rows.Scan(&m.ID, &m.Team1ID, &m.Team2ID, &m.WinningTeamID, &m.StartTime, &m.EndTime, &m.Location, &m.Notes)
		matches = append(matches, m)
	}

	c.JSON(http.StatusOK, matches)
}
