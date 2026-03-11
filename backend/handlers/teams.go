package handlers

import (
	"database/sql"
	"net/http"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
)

// CreateTeam creates a new team and assigns the creator to it
// POST /api/teams
func CreateTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

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

		// Assign the creator's player profile to this team
		db.Exec(`UPDATE players SET team_id = ? WHERE user_id = ?`, teamID, userID)

		c.JSON(http.StatusCreated, gin.H{"id": teamID, "name": req.Name})
	}
}

// GetTeams returns all teams with member count
// GET /api/teams
func GetTeams(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT t.id, t.name, t.created_at, COUNT(p.id) as member_count
			FROM teams t
			LEFT JOIN players p ON p.team_id = t.id
			GROUP BY t.id
			ORDER BY t.name
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
			return
		}
		defer rows.Close()

		type TeamWithCount struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			MemberCount int    `json:"member_count"`
		}

		teams := []TeamWithCount{}
		for rows.Next() {
			var t TeamWithCount
			var createdAt string
			if err := rows.Scan(&t.ID, &t.Name, &createdAt, &t.MemberCount); err == nil {
				teams = append(teams, t)
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

// GetTeamMembers returns players on a team
// GET /api/teams/:id/members
func GetTeamMembers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamID := c.Param("id")

		rows, err := db.Query(`
			SELECT p.id, p.user_id, p.name
			FROM players p
			WHERE p.team_id = ?
			ORDER BY p.name
		`, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team members"})
			return
		}
		defer rows.Close()

		type TeamMember struct {
			PlayerID int    `json:"player_id"`
			UserID   int    `json:"user_id"`
			Name     string `json:"name"`
		}

		members := []TeamMember{}
		for rows.Next() {
			var m TeamMember
			if rows.Scan(&m.PlayerID, &m.UserID, &m.Name) == nil {
				members = append(members, m)
			}
		}

		c.JSON(http.StatusOK, members)
	}
}

// JoinTeam assigns the current user's player to a team
// POST /api/teams/:id/join
func JoinTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		teamID := c.Param("id")

		var teamName string
		if err := db.QueryRow(`SELECT name FROM teams WHERE id = ?`, teamID).Scan(&teamName); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}

		// Check team isn't full (max 2 players)
		var memberCount int
		db.QueryRow(`SELECT COUNT(*) FROM players WHERE team_id = ?`, teamID).Scan(&memberCount)
		if memberCount >= 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Team is full"})
			return
		}

		_, err := db.Exec(`UPDATE players SET team_id = ? WHERE user_id = ?`, teamID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined team", "team_name": teamName})
	}
}

// LeaveTeam removes the current user's player from their team
// POST /api/teams/leave
func LeaveTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		_, err := db.Exec(`UPDATE players SET team_id = NULL WHERE user_id = ?`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Left team"})
	}
}
