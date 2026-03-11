package handlers

import (
	"database/sql"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateTeam creates a new team
// POST /api/teams
func CreateTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := db.Exec(`INSERT INTO teams (name) VALUES (?)`, req.Name)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Team name already exists"})
			return
		}

		teamID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"id": teamID, "name": req.Name})
	}
}

// GetTeams returns all teams
// GET /api/teams
func GetTeams(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, created_at FROM teams ORDER BY name`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
			return
		}
		defer rows.Close()

		teams := []models.Team{}
		for rows.Next() {
			var team models.Team
			if err := rows.Scan(&team.ID, &team.Name, &team.CreatedAt); err == nil {
				teams = append(teams, team)
			}
		}

		c.JSON(http.StatusOK, teams)
	}
}

// GetTeam returns a single team by ID
// GET /api/teams/:id
func GetTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamID := c.Param("id")

		var team models.Team
		err := db.QueryRow(`
			SELECT id, name, created_at
			FROM teams
			WHERE id = ?
		`, teamID).Scan(&team.ID, &team.Name, &team.CreatedAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team"})
			return
		}

		c.JSON(http.StatusOK, team)
	}
}
