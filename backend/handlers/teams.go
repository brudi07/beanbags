package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CreateTeam creates a new team and adds the creator as its first member
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

		result, err := db.Exec(`INSERT INTO teams (name, is_managed) VALUES (?, 1)`, req.Name)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Team name already exists"})
			return
		}

		teamID, _ := result.LastInsertId()

		// Add the creator to team_members
		var playerID int
		if err := db.QueryRow(`SELECT id FROM players WHERE user_id = ?`, userID).Scan(&playerID); err == nil {
			db.Exec(`INSERT OR IGNORE INTO team_members (team_id, player_id) VALUES (?, ?)`, teamID, playerID)
		}

		c.JSON(http.StatusCreated, gin.H{"id": teamID, "name": req.Name})
	}
}

// GetTeams returns all managed teams with member count and interested leagues
// GET /api/teams
func GetTeams(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT t.id, t.name,
			    (SELECT COUNT(*) FROM team_members WHERE team_id = t.id) as member_count,
			    COALESCE(GROUP_CONCAT(l.id || ':::' || l.name, '|||'), '') as interested_leagues
			FROM teams t
			LEFT JOIN team_league_interests tli ON tli.team_id = t.id
			LEFT JOIN leagues l ON l.id = tli.league_id
			WHERE t.is_managed = 1
			GROUP BY t.id
			ORDER BY t.name
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
			return
		}
		defer rows.Close()

		type InterestedLeague struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		type TeamWithCount struct {
			ID                int                `json:"id"`
			Name              string             `json:"name"`
			MemberCount       int                `json:"member_count"`
			InterestedLeagues []InterestedLeague `json:"interested_leagues"`
		}

		teams := []TeamWithCount{}
		for rows.Next() {
			var t TeamWithCount
			var interestedLeaguesStr string
			if err := rows.Scan(&t.ID, &t.Name, &t.MemberCount, &interestedLeaguesStr); err == nil {
				t.InterestedLeagues = []InterestedLeague{}
				if interestedLeaguesStr != "" {
					for _, part := range strings.Split(interestedLeaguesStr, "|||") {
						pieces := strings.SplitN(part, ":::", 2)
						if len(pieces) == 2 {
							id, err := strconv.Atoi(pieces[0])
							if err == nil {
								t.InterestedLeagues = append(t.InterestedLeagues, InterestedLeague{ID: id, Name: pieces[1]})
							}
						}
					}
				}
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

		var id int
		var name string
		err := db.QueryRow(`SELECT id, name FROM teams WHERE id = ?`, teamID).Scan(&id, &name)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id, "name": name})
	}
}

// GetTeamMembers returns players on a team
// GET /api/teams/:id/members
func GetTeamMembers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		teamID := c.Param("id")

		rows, err := db.Query(`
			SELECT p.id, p.user_id, p.name
			FROM team_members tm
			JOIN players p ON tm.player_id = p.id
			WHERE tm.team_id = ?
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

// JoinTeam adds the current user to a team (max 2 members)
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
		if err := db.QueryRow(`SELECT name FROM teams WHERE id = ? AND is_managed = 1`, teamID).Scan(&teamName); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
			return
		}

		// Check team isn't full
		var memberCount int
		db.QueryRow(`SELECT COUNT(*) FROM team_members WHERE team_id = ?`, teamID).Scan(&memberCount)
		if memberCount >= 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Team is full"})
			return
		}

		// Get the user's player profile
		var playerID int
		if err := db.QueryRow(`SELECT id FROM players WHERE user_id = ?`, userID).Scan(&playerID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player profile not found"})
			return
		}

		_, err := db.Exec(`INSERT OR IGNORE INTO team_members (team_id, player_id) VALUES (?, ?)`, teamID, playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join team"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Joined team", "team_name": teamName})
	}
}

// LeaveTeam removes the current user from a specific team
// POST /api/teams/leave
func LeaveTeam(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var req struct {
			TeamID int `json:"team_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "team_id is required"})
			return
		}

		var playerID int
		if err := db.QueryRow(`SELECT id FROM players WHERE user_id = ?`, userID).Scan(&playerID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Player profile not found"})
			return
		}

		_, err := db.Exec(`DELETE FROM team_members WHERE team_id = ? AND player_id = ?`, req.TeamID, playerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave team"})
			return
		}

		// Delete the team if it now has no members
		var remaining int
		db.QueryRow(`SELECT COUNT(*) FROM team_members WHERE team_id = ?`, req.TeamID).Scan(&remaining)
		if remaining == 0 {
			db.Exec(`DELETE FROM teams WHERE id = ?`, req.TeamID)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Left team"})
	}
}

// AddTeamInterest signals that a team is interested in a league
// POST /api/teams/:id/interests
func AddTeamInterest(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		teamID := c.Param("id")

		// Verify caller is a member of the team
		var count int
		db.QueryRow(`
			SELECT COUNT(*) FROM team_members tm
			JOIN players p ON tm.player_id = p.id
			WHERE tm.team_id = ? AND p.user_id = ?
		`, teamID, userID).Scan(&count)
		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this team"})
			return
		}

		var req struct {
			LeagueID int `json:"league_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.Exec(`INSERT OR IGNORE INTO team_league_interests (team_id, league_id) VALUES (?, ?)`, teamID, req.LeagueID)

		c.JSON(http.StatusOK, gin.H{"message": "Interest added"})
	}
}

// RemoveTeamInterest removes a team's interest in a league
// DELETE /api/teams/:id/interests/:leagueId
func RemoveTeamInterest(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetCurrentUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		teamID := c.Param("id")
		leagueID := c.Param("leagueId")

		// Verify caller is a member of the team
		var count int
		db.QueryRow(`
			SELECT COUNT(*) FROM team_members tm
			JOIN players p ON tm.player_id = p.id
			WHERE tm.team_id = ? AND p.user_id = ?
		`, teamID, userID).Scan(&count)
		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this team"})
			return
		}

		db.Exec(`DELETE FROM team_league_interests WHERE team_id = ? AND league_id = ?`, teamID, leagueID)

		c.JSON(http.StatusOK, gin.H{"message": "Interest removed"})
	}
}
