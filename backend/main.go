package main

import (
	"log"
	"net/http"

	"brudi07/beanbags/database"
	"brudi07/beanbags/handlers"

	_ "modernc.org/sqlite"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db, err := database.InitializeDatabase("./data/beanbags.db")
	if err != nil {
		log.Fatal("failed to initialize database:", err)
	}
	defer db.Close()

	r := gin.Default()

	// CORS configuration for frontend
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // Add your frontend URLs
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// =========================
	// Public Routes (No Auth)
	// =========================

	// Authentication
	r.POST("/api/auth/signup", handlers.Register(db))
	r.POST("/api/auth/login", handlers.Login(db))
	r.POST("/api/auth/forgot-password", func(c *gin.Context) {
		// TODO: Implement password reset
		c.JSON(http.StatusNotImplemented, gin.H{"message": "Password reset not yet implemented"})
	})

	// =========================
	// Protected Routes (Auth Required)
	// =========================

	authorized := r.Group("/api")
	authorized.Use(handlers.AuthMiddleware())
	{
		// Leagues - Public Browse
		authorized.GET("/leagues/public", handlers.GetPublicLeagues(db))
		authorized.GET("/leagues/my-leagues", handlers.GetMyLeagues(db))
		authorized.GET("/leagues/:id", handlers.GetLeagueByID(db))
		authorized.GET("/leagues/:id/members", handlers.GetLeagueMembers(db))
		authorized.GET("/leagues/:id/schedule", handlers.GetLeagueSchedule(db))
		authorized.GET("/leagues/:id/standings", handlers.GetLeagueStandings(db))

		// Leagues - Member Actions
		authorized.POST("/leagues/:id/join", handlers.JoinLeague(db))
		authorized.POST("/leagues/:id/leave", handlers.LeaveLeague(db))

		// Leagues - Organizer Actions
		authorized.POST("/leagues", handlers.CreateLeague(db))
		authorized.PATCH("/leagues/:id", handlers.UpdateLeague(db))
		authorized.PATCH("/leagues/:id/games/:gameId/reschedule", handlers.RescheduleLeagueGame(db))
		authorized.POST("/leagues/:id/generate-schedule", handlers.GenerateLeagueSchedule(db))

		// Games/Matches
		authorized.POST("/games", handlers.CreateMatch(db))
		authorized.GET("/games/:id", handlers.GetMatch(db))
		authorized.POST("/games/:id/complete", handlers.CompleteGame(db))
		authorized.GET("/games/:id/results", handlers.GetGameResults(db))

		// Teams & Players
		authorized.POST("/teams", handlers.CreateTeam(db))
		authorized.GET("/teams", handlers.GetTeams(db))
		authorized.POST("/players", handlers.CreatePlayer(db))
		authorized.GET("/players", handlers.GetPlayers(db))
		authorized.GET("/players/:id", handlers.GetPlayer(db))
		authorized.GET("/players/:id/stats", handlers.GetPlayerStats(db))
		authorized.GET("/players/:id/heatmap", handlers.GetPlayerHeatmap(db))

		// Rounds & Throws
		authorized.POST("/rounds", handlers.CreateRound(db))
		authorized.POST("/rounds/:id/throws", handlers.AddThrow(db))
	}

	log.Println("🚀 Server starting on :8080")
	r.Run(":8080")
}
