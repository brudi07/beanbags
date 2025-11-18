package handlers

import (
	"brudi07/beanbags/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CreatePlayer(c *gin.Context) {
	var p models.Player
	if err := c.BindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.DB.Exec(`INSERT INTO players (name, team_id) VALUES (?, ?)`, p.Name, p.TeamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	p.ID = int(id)
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) GetPlayers(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, name, team_id, created_at FROM players`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var p models.Player
		rows.Scan(&p.ID, &p.Name, &p.TeamID, &p.CreatedAt)
		players = append(players, p)
	}

	c.JSON(http.StatusOK, players)
}
