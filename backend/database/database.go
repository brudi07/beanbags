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

// runSchema executes the schema statements if tables don’t exist.
func runSchema(db *sql.DB) error {
	schema := `
	PRAGMA foreign_keys = ON;

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
	    name            TEXT NOT NULL,
	    team_id         INTEGER,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (team_id) REFERENCES teams (id)
	);

	-- =========================
	-- Leagues and Tournaments
	-- =========================

	CREATE TABLE IF NOT EXISTS leagues (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		start_date TEXT,
		end_date TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		bracket_round INTEGER,      -- for tournament bracket structure
		bracket_position INTEGER    -- position within round (for visualization)
	);

	CREATE TABLE IF NOT EXISTS rounds (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    match_id        INTEGER NOT NULL,
	    round_number    INTEGER NOT NULL,
	    team1_score     INTEGER DEFAULT 0,
	    team2_score     INTEGER DEFAULT 0,
	    ended_in_bust   BOOLEAN DEFAULT 0,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (match_id) REFERENCES matches (id)
	);

	CREATE TABLE IF NOT EXISTS throws (
	    id              INTEGER PRIMARY KEY AUTOINCREMENT,
	    round_id        INTEGER NOT NULL,
	    player_id       INTEGER NOT NULL,
	    throw_number    INTEGER NOT NULL,
	    throw_type      TEXT CHECK(throw_type IN (
	                            'in_hole',
	                            'on_board',
	                            'off_board',
	                            'intentional_off'
	                        )) NOT NULL,
	    x_position      REAL,
	    y_position      REAL,
	    points_earned   INTEGER DEFAULT 0,
	    caused_bust     BOOLEAN DEFAULT 0,
	    auto_thrown_off BOOLEAN DEFAULT 0,
	    created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (round_id) REFERENCES rounds (id),
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
	    bags_in_hole            INTEGER DEFAULT 0,
	    bags_on_board           INTEGER DEFAULT 0,
	    bags_off_board          INTEGER DEFAULT 0,
	    intentional_offs        INTEGER DEFAULT 0,
	    busts_caused            INTEGER DEFAULT 0,
	    points_total            INTEGER DEFAULT 0,
	    average_points_per_round REAL DEFAULT 0,
	    updated_at              DATETIME DEFAULT CURRENT_TIMESTAMP,
	    FOREIGN KEY (player_id) REFERENCES players (id),
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
		position_in_next INTEGER,         -- 1 = winner to team1, 2 = winner to team2
		loser_next_match_id INTEGER REFERENCES matches(id), -- for double elimination
		position_in_loser_next INTEGER,   -- where the loser goes
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	// Split and execute statements (easier to debug if one fails)
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
