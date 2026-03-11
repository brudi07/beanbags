package models

import "time"

// =========================
// User and Authentication
// =========================

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Roles        []string  `json:"roles"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RegisterRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	Password  string   `json:"password" binding:"required,min=6"`
	FirstName string   `json:"first_name" binding:"required"`
	LastName  string   `json:"last_name" binding:"required"`
	Roles     []string `json:"roles" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// =========================
// Core Entities
// =========================

type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Player struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	TeamID    *int      `json:"team_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// =========================
// Leagues
// =========================

type League struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	OrganizerID   int       `json:"organizer_id"`
	OrganizerName string    `json:"organizer_name"`
	Format        string    `json:"format"`
	GamesPerMatch int       `json:"games_per_match"`
	MaxTeams      int       `json:"max_teams"`
	CurrentTeams  int       `json:"current_teams"`
	StartDate     string    `json:"start_date"`
	WeeksOfPlay   int       `json:"weeks_of_play"`
	Location      string    `json:"location"`
	IsPublic      bool      `json:"is_public"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateLeagueRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	Format        string `json:"format" binding:"required,oneof=1v1 2v2"`
	GamesPerMatch int    `json:"games_per_match" binding:"required,oneof=1 3 5"`
	MaxTeams      int    `json:"max_teams" binding:"required,min=2,max=64"`
	StartDate     string `json:"start_date" binding:"required"`
	WeeksOfPlay   int    `json:"weeks_of_play" binding:"required,min=1,max=52"`
	Location      string `json:"location" binding:"required"`
	IsPublic      bool   `json:"is_public"`
}

type LeagueMember struct {
	LeagueID   int       `json:"league_id"`
	PlayerID   int       `json:"player_id"`
	PlayerName string    `json:"player_name"`
	JoinedAt   time.Time `json:"joined_at"`
}

type LeagueGame struct {
	ID             int       `json:"id"`
	LeagueID       int       `json:"league_id"`
	MatchNumber    int       `json:"match_number"`
	GameNumber     int       `json:"game_number"`
	ScheduledDate  string    `json:"scheduled_date"`
	Team1          string    `json:"team1"`
	Team2          string    `json:"team2"`
	Team1PlayerIDs []int     `json:"team1_player_ids"`
	Team2PlayerIDs []int     `json:"team2_player_ids"`
	Status         string    `json:"status"`
	WinningTeam    *int      `json:"winning_team,omitempty"`
	Team1Score     *int      `json:"team1_score,omitempty"`
	Team2Score     *int      `json:"team2_score,omitempty"`
	GameID         *int      `json:"game_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LeagueSchedule struct {
	Date  string       `json:"date"`
	Games []LeagueGame `json:"games"`
}

type LeagueStanding struct {
	LeagueID      int     `json:"league_id"`
	TeamID        int     `json:"team_id"`
	TeamName      string  `json:"team_name"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	Ties          int     `json:"ties"`
	PointsFor     int     `json:"points_for"`
	PointsAgainst int     `json:"points_against"`
	PointDiff     int     `json:"point_diff"`
	WinPercentage float64 `json:"win_percentage"`
}

// =========================
// Matches and Games
// =========================

type Match struct {
	ID            int        `json:"id"`
	LeagueID      *int       `json:"league_id,omitempty"`
	TournamentID  *int       `json:"tournament_id,omitempty"`
	Team1ID       int        `json:"team1_id"`
	Team2ID       int        `json:"team2_id"`
	WinningTeamID *int       `json:"winning_team_id,omitempty"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       *time.Time `json:"end_time,omitempty"`
	Location      string     `json:"location"`
	Notes         string     `json:"notes"`
	Status        string     `json:"status"`
}

type Round struct {
	ID          int       `json:"id"`
	MatchID     int       `json:"match_id"`
	RoundNumber int       `json:"round_number"`
	Team1Score  int       `json:"team1_score"`
	Team2Score  int       `json:"team2_score"`
	Team1Points int       `json:"team1_points"`
	Team2Points int       `json:"team2_points"`
	Team1Busted bool      `json:"team1_busted"`
	Team2Busted bool      `json:"team2_busted"`
	Throws      []Throw   `json:"throws,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Throw struct {
	ID           int       `json:"id"`
	RoundID      int       `json:"round_id"`
	PlayerID     int       `json:"player_id"`
	Team         int       `json:"team"`
	ThrowNumber  int       `json:"throw_number"`
	Result       string    `json:"result"`
	XPosition    float64   `json:"x_position"`
	YPosition    float64   `json:"y_position"`
	Rotation     float64   `json:"rotation"`
	PointsEarned int       `json:"points_earned"`
	CreatedAt    time.Time `json:"created_at"`
}

type RoundResult struct {
	Round       int     `json:"round"`
	Throws      []Throw `json:"throws"`
	Team1Points int     `json:"team1_points"`
	Team2Points int     `json:"team2_points"`
	Team1Busted bool    `json:"team1_busted,omitempty"`
	Team2Busted bool    `json:"team2_busted,omitempty"`
}

type GameSubmitRequest struct {
	Winner       int           `json:"winner"`
	FinalScore   FinalScore    `json:"final_score"`
	TotalRounds  int           `json:"total_rounds"`
	RoundHistory []RoundResult `json:"round_history"`
	CompletedAt  string        `json:"completed_at"`
}

type FinalScore struct {
	Team1 int `json:"team1"`
	Team2 int `json:"team2"`
}

// =========================
// Tournaments
// =========================

type Tournament struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	LeagueID  *int      `json:"league_id,omitempty"`
	Format    string    `json:"format"`
	Location  string    `json:"location"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
	TeamIDs   []int     `json:"team_ids,omitempty"`
}

type TournamentInput struct {
	Name     string `json:"name" binding:"required"`
	LeagueID *int   `json:"league_id"`
	Format   string `json:"format" binding:"required"`
	TeamIDs  []int  `json:"team_ids" binding:"required"`
}

type TournamentMatch struct {
	MatchID  int    `json:"match_id"`
	Round    int    `json:"round"`
	Team1ID  int    `json:"team1_id"`
	Team2ID  int    `json:"team2_id"`
	WinnerID *int   `json:"winner_id,omitempty"`
	Location string `json:"location"`
}
