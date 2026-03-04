package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type League struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Location  string `json:"location"`
}

type LeagueStanding struct {
	TeamID        int    `json:"team_id"`
	TeamName      string `json:"team_name"`
	Wins          int    `json:"wins"`
	Losses        int    `json:"losses"`
	Ties          int    `json:"ties"`
	PointsFor     int    `json:"points_for"`
	PointsAgainst int    `json:"points_against"`
	PointDiff     int    `json:"point_diff"`
}

// CreateLeague handles POST /leagues
func CreateLeague(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var league League
		if err := c.ShouldBindJSON(&league); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := db.Exec(`
			INSERT INTO leagues (name, start_date, end_date, location)
			VALUES (?, ?, ?, ?)`,
			league.Name, league.StartDate, league.EndDate, league.Location)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id, _ := res.LastInsertId()
		league.ID = int(id)

		c.JSON(http.StatusCreated, league)
	}
}

// GetLeagues handles GET /leagues
func GetLeagues(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, start_date, end_date, location FROM leagues`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var leagues []League
		for rows.Next() {
			var l League
			if err := rows.Scan(&l.ID, &l.Name, &l.StartDate, &l.EndDate, &l.Location); err == nil {
				leagues = append(leagues, l)
			}
		}

		c.JSON(http.StatusOK, leagues)
	}
}

// GetLeagueStandings handles GET /leagues/:id/standings
func GetLeagueStandings(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		leagueID := c.Param("id")

		rows, err := db.Query(`
			SELECT ls.team_id, t.name, ls.wins, ls.losses, ls.ties,
				   ls.points_for, ls.points_against, ls.point_diff
			FROM league_standings ls
			JOIN teams t ON ls.team_id = t.id
			WHERE ls.league_id = ?
			ORDER BY ls.wins DESC, ls.point_diff DESC`, leagueID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var standings []LeagueStanding
		for rows.Next() {
			var s LeagueStanding
			if err := rows.Scan(&s.TeamID, &s.TeamName, &s.Wins, &s.Losses, &s.Ties, &s.PointsFor, &s.PointsAgainst, &s.PointDiff); err == nil {
				standings = append(standings, s)
			}
		}

		c.JSON(http.StatusOK, standings)
	}
}
