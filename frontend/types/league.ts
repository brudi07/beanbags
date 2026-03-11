export interface League {
    id: number
    name: string
    description: string
    organizer_id: number
    organizer_name: string
    status: 'active' | 'completed' | 'upcoming'
    format: '1v1' | '2v2'
    games_per_match: 1 | 3 | 5
    max_teams: number
    current_teams: number
    start_date: string
    weeks_of_play: number
    location: string
    is_public: boolean
    created_at: string
    updated_at: string
}

export interface LeaguePlayer {
    league_id: number
    player_id: number
    player_name: string
    joined_at: string
    role: 'player' | 'organizer'
}

export interface CreateLeagueData {
    name: string
    description: string
    format: '1v1' | '2v2'
    games_per_match: 1 | 3 | 5
    max_teams: number
    start_date: string
    weeks_of_play: number
    location: string
    is_public: boolean
}

export interface LeagueGame {
    id: number
    league_id: number
    match_number: number
    game_number: number
    scheduled_date: string
    team1: string  // ✅ Team name (e.g., "Ben & Sarah")
    team2: string  // ✅ Team name (e.g., "Alex & Jordan")
    team1_player_ids: number[]  // Player IDs for team 1
    team2_player_ids: number[]  // Player IDs for team 2
    status: 'scheduled' | 'in_progress' | 'completed'
    winning_team?: 1 | 2  // Which team won (1 or 2)
    team1_score?: number
    team2_score?: number
    game_id?: number
    created_at: string
    updated_at: string
}

export interface LeagueSchedule {
    date: string
    games: LeagueGame[]
}