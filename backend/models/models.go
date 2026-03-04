package models

import "time"

type Team struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Player struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TeamID    int       `json:"team_id"`
	CreatedAt time.Time `json:"created_at"`
}

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
}

type Round struct {
	ID          int       `json:"id"`
	MatchID     int       `json:"match_id"`
	RoundNumber int       `json:"round_number"`
	Team1Score  int       `json:"team1_score"`
	Team2Score  int       `json:"team2_score"`
	EndedInBust bool      `json:"ended_in_bust"`
	CreatedAt   time.Time `json:"created_at"`
}

type Throw struct {
	ID            int       `json:"id"`
	RoundID       int       `json:"round_id"`
	PlayerID      int       `json:"player_id"`
	ThrowNumber   int       `json:"throw_number"`
	ThrowType     string    `json:"throw_type"`
	XPosition     float64   `json:"x_position"`
	YPosition     float64   `json:"y_position"`
	PointsEarned  int       `json:"points_earned"`
	CausedBust    bool      `json:"caused_bust"`
	AutoThrownOff bool      `json:"auto_thrown_off"`
	CreatedAt     time.Time `json:"created_at"`
}

type League struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	CreatedAt time.Time `json:"created_at"`
}

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

type LeagueStanding struct {
	TeamID   int    `json:"team_id"`
	TeamName string `json:"team_name"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
}

type TournamentMatch struct {
	MatchID  int    `json:"match_id"`
	Round    int    `json:"round"`
	Team1ID  int    `json:"team1_id"`
	Team2ID  int    `json:"team2_id"`
	WinnerID *int   `json:"winner_id,omitempty"`
	Location string `json:"location"`
}
