package handlers

import (
	"brudi07/beanbags/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateTeam(c *gin.Context) {
	var team models.Team
	if err := c.BindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.DB.Exec(`INSERT INTO teams (name) VALUES (?)`, team.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	team.ID = int(id)
	c.JSON(http.StatusCreated, team)
}

func (h *Handler) GetTeams(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, name, created_at FROM teams`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var teams []models.Team
	for rows.Next() {
		var t models.Team
		rows.Scan(&t.ID, &t.Name, &t.CreatedAt)
		teams = append(teams, t)
	}

	c.JSON(http.StatusOK, teams)
}
