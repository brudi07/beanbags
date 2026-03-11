package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"brudi07/beanbags/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("your-secret-key-change-this-in-production") // TODO: Move to environment variable

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// Register creates a new user account
func Register(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Start transaction
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		defer tx.Rollback()

		// Create user
		result, err := tx.Exec(`
			INSERT INTO users (email, password_hash, first_name, last_name)
			VALUES (?, ?, ?, ?)
		`, req.Email, string(hashedPassword), req.FirstName, req.LastName)

		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "An account with this email already exists"})
			return
		}

		userID, _ := result.LastInsertId()

		// Insert roles
		for _, role := range req.Roles {
			_, err = tx.Exec(`
				INSERT INTO user_roles (user_id, role)
				VALUES (?, ?)
			`, userID, role)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign roles"})
				return
			}
		}

		// Create player entry if user has player role
		hasPlayerRole := false
		for _, role := range req.Roles {
			if role == "player" {
				hasPlayerRole = true
				break
			}
		}

		if hasPlayerRole {
			playerName := req.FirstName + " " + req.LastName
			_, err = tx.Exec(`
				INSERT INTO players (user_id, name)
				VALUES (?, ?)
			`, userID, playerName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player profile"})
				return
			}
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
			return
		}

		// Generate JWT token
		token, err := generateToken(int(userID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Return user data
		user := models.User{
			ID:        int(userID),
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Roles:     req.Roles,
		}

		c.JSON(http.StatusCreated, models.AuthResponse{
			Token: token,
			User:  user,
		})
	}
}

// Login authenticates a user
func Login(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get user from database
		var user models.User
		var passwordHash string

		err := db.QueryRow(`
			SELECT id, email, password_hash, first_name, last_name, created_at, updated_at
			FROM users
			WHERE email = ?
		`, req.Email).Scan(
			&user.ID,
			&user.Email,
			&passwordHash,
			&user.FirstName,
			&user.LastName,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Get user roles
		rows, err := db.Query(`
			SELECT role FROM user_roles WHERE user_id = ?
		`, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
			return
		}
		defer rows.Close()

		user.Roles = []string{}
		for rows.Next() {
			var role string
			if err := rows.Scan(&role); err != nil {
				continue
			}
			user.Roles = append(user.Roles, role)
		}

		// Generate JWT token
		token, err := generateToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, models.AuthResponse{
			Token: token,
			User:  user,
		})
	}
}

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims and store user ID in context
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("user_id", claims.UserID)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
}

// Helper function to generate JWT token
func generateToken(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// GetCurrentUserID helper extracts user ID from context
func GetCurrentUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(int)
	return id, ok
}
