package main

import (
	"log"
	"net/http"

	"brudi07/beanbags/database"
	"brudi07/beanbags/handlers"

	_ "modernc.org/sqlite"

	"github.com/gin-gonic/gin"
)

func main() {
	// ✅ Initialize and migrate DB
	db, err := database.InitializeDatabase("./beanbags.db")
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	h := handlers.NewHandler(db)
	r := gin.Default()

	r.POST("/teams", h.CreateTeam)
	r.GET("/teams", h.GetTeams)

	r.POST("/players", h.CreatePlayer)
	r.GET("/players", h.GetPlayers)

	r.POST("/matches", h.CreateMatch)
	r.GET("/matches", h.GetMatches)

	// Throws
	r.POST("/rounds/:id/throws", h.AddThrow)
	r.GET("/players/:id/heatmap", h.GetPlayerHeatmap)
	r.POST("/rounds", h.CreateRound)

	// Leagues
	r.POST("/leagues", handlers.CreateLeague(db))
	r.GET("/leagues", handlers.GetLeagues(db))
	r.GET("/leagues/:id/standings", handlers.GetLeagueStandings(db))

	// Tournaments
	r.POST("/tournaments", handlers.CreateTournament(db))
	r.GET("/tournaments", handlers.GetTournaments(db))
	r.POST("/tournaments/:id/bracket", handlers.GenerateBracket(db))
	r.GET("/tournaments/:id/bracket", handlers.GetBracket(db))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Run(":8080") // start server
}
