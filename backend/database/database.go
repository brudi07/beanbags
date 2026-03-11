package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// InitializeDatabase opens (or creates) the SQLite DB and runs schema creation if needed.
func InitializeDatabase(dbPath string) (*sql.DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll("./data", os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enforce foreign key integrity
	if _, err = db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Run schema initialization
	if err = runSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("✅ SQLite database initialized successfully.")
	return db, nil
}

// runSchema executes the schema statements if tables don't exist.
func runSchema(db *sql.DB) error {
	schema := `
	PRAGMA foreign_keys = ON;

	-- =========================
	-- Users and Authentication
	-- =========================

	CREATE TABLE IF NOT EXISTS users (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    email           TEXT NOT NULL UNIQUE,
	    password_hash   TEXT NOT NULL,
	    first_name      TEXT NOT NULL DEFAULT '',
	    last_name       TEXT NOT NULL DEFAULT '',
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_roles (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    user_id         INTEGER NOT NULL,
	    role            TEXT CHECK(role IN ('player', 'organizer')) NOT NULL,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
	    UNIQUE(user_id, role)
	);

	-- =========================
	-- Core Entities
	-- =========================

	CREATE TABLE IF NOT EXISTS teams (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    name            TEXT NOT NULL UNIQUE,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS players (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    user_id         INTEGER NOT NULL,
	    name            TEXT NOT NULL,
	    team_id         INTEGER,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
	    FOREIGN KEY (team_id) REFERENCES teams (id),
	    UNIQUE(user_id)
	);

	-- =========================
	-- Leagues and Tournaments
	-- =========================

	CREATE TABLE IF NOT EXISTS leagues (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		name              TEXT NOT NULL,
		description       TEXT,
		organizer_id      INTEGER NOT NULL,
		format            TEXT CHECK(format IN ('1v1', '2v2')) DEFAULT '2v2',
		games_per_match   INTEGER CHECK(games_per_match IN (1, 3, 5)) DEFAULT 3,
		max_teams         INTEGER DEFAULT 8,
		current_teams     INTEGER DEFAULT 0,
		start_date        TEXT NOT NULL,
		weeks_of_play     INTEGER DEFAULT 8,
		location          TEXT NOT NULL,
		is_public         BOOLEAN DEFAULT 1,
		status            TEXT CHECK(status IN ('upcoming', 'active', 'completed')) DEFAULT 'upcoming',
		created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (organizer_id) REFERENCES users (id)
	);

	CREATE TABLE IF NOT EXISTS league_members (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id       INTEGER NOT NULL,
		player_id       INTEGER NOT NULL,
		joined_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (league_id) REFERENCES leagues (id) ON DELETE CASCADE,
		FOREIGN KEY (player_id) REFERENCES players (id) ON DELETE CASCADE,
		UNIQUE(league_id, player_id)
	);

	CREATE TABLE IF NOT EXISTS league_games (
		id                INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id         INTEGER NOT NULL,
		match_number      INTEGER NOT NULL,
		game_number       INTEGER NOT NULL,
		scheduled_date    TEXT NOT NULL,
		team1             TEXT NOT NULL,
		team2             TEXT NOT NULL,
		team1_player_ids  TEXT NOT NULL,
		team2_player_ids  TEXT NOT NULL,
		status            TEXT CHECK(status IN ('scheduled', 'in_progress', 'completed')) DEFAULT 'scheduled',
		winning_team      INTEGER CHECK(winning_team IN (1, 2)),
		game_id           INTEGER,
		created_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at        DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (league_id) REFERENCES leagues (id) ON DELETE CASCADE,
		FOREIGN KEY (game_id) REFERENCES matches (id)
	);

	CREATE TABLE IF NOT EXISTS tournaments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id INTEGER REFERENCES leagues(id) ON DELETE SET NULL,
		name TEXT NOT NULL,
		location TEXT,
		start_date TEXT,
		end_date TEXT,
		format TEXT CHECK(format IN ('single_elimination', 'double_elimination')) DEFAULT 'single_elimination',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tournament_teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tournament_id INTEGER NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
		team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
		UNIQUE(tournament_id, team_id)
	);

	-- =========================
	-- Matches, Rounds, and Throws
	-- =========================

	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id INTEGER REFERENCES leagues(id) ON DELETE SET NULL,
		tournament_id INTEGER REFERENCES tournaments(id) ON DELETE SET NULL,
		team1_id INTEGER REFERENCES teams(id),
		team2_id INTEGER REFERENCES teams(id),
		winning_team_id INTEGER REFERENCES teams(id),
		start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
		end_time DATETIME,
		location TEXT,
		notes TEXT,
		bracket_round INTEGER,
		bracket_position INTEGER,
		status TEXT CHECK(status IN ('pending', 'active', 'completed', 'abandoned')) DEFAULT 'pending'
	);

	CREATE TABLE IF NOT EXISTS rounds (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    match_id        INTEGER NOT NULL,
	    round_number    INTEGER NOT NULL,
	    team1_score     INTEGER DEFAULT 0,
	    team2_score     INTEGER DEFAULT 0,
	    team1_points    INTEGER DEFAULT 0,
	    team2_points    INTEGER DEFAULT 0,
	    team1_busted    BOOLEAN DEFAULT 0,
	    team2_busted    BOOLEAN DEFAULT 0,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (match_id) REFERENCES matches (id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS throws (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    round_id        INTEGER NOT NULL,
	    player_id       INTEGER NOT NULL,
	    team            INTEGER CHECK(team IN (1, 2)) NOT NULL,
	    throw_number    INTEGER NOT NULL,
	    result          TEXT CHECK(result IN ('hole', 'board', 'miss', 'ito')) NOT NULL,
	    x_position      REAL,
	    y_position      REAL,
	    rotation        REAL DEFAULT 0,
	    points_earned   INTEGER DEFAULT 0,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (round_id) REFERENCES rounds (id) ON DELETE CASCADE,
	    FOREIGN KEY (player_id) REFERENCES players (id)
	);

	-- =========================
	-- Stats and Standings
	-- =========================

	CREATE TABLE IF NOT EXISTS player_stats (
	    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
	    player_id               INTEGER NOT NULL,
	    league_id               INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
	    total_throws            INTEGER DEFAULT 0,
	    holes                   INTEGER DEFAULT 0,
	    boards                  INTEGER DEFAULT 0,
	    misses                  INTEGER DEFAULT 0,
	    itos                    INTEGER DEFAULT 0,
	    busts                   INTEGER DEFAULT 0,
	    points_contributed      INTEGER DEFAULT 0,
	    accuracy                INTEGER DEFAULT 0,
	    points_per_round        REAL DEFAULT 0,
	    differential_per_round  REAL DEFAULT 0,
	    updated_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (player_id) REFERENCES players (id) ON DELETE CASCADE,
	    UNIQUE(player_id, league_id)
	);

	CREATE TABLE IF NOT EXISTS league_standings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
		team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
		wins INTEGER DEFAULT 0,
		losses INTEGER DEFAULT 0,
		ties INTEGER DEFAULT 0,
		points_for INTEGER DEFAULT 0,
		points_against INTEGER DEFAULT 0,
		point_diff INTEGER DEFAULT 0,
		win_percentage REAL DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(league_id, team_id)
	);

	-- =========================
	-- Tournament Brackets
	-- =========================

	CREATE TABLE IF NOT EXISTS tournament_brackets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tournament_id INTEGER REFERENCES tournaments(id) ON DELETE CASCADE,
		round_number INTEGER NOT NULL,
		match_id INTEGER REFERENCES matches(id) ON DELETE CASCADE,
		next_match_id INTEGER REFERENCES matches(id),
		position_in_next INTEGER,
		loser_next_match_id INTEGER REFERENCES matches(id),
		position_in_loser_next INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Split and execute statements
	statements := strings.Split(schema, ";")
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("error executing schema statement %q: %w", stmt, err)
		}
	}

return nil
}
