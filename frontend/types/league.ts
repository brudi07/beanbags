export interface League {
    id: string
    name: string
    description: string
    organizerId: string
    organizerName: string
    status: 'active' | 'completed' | 'upcoming'
    format: '1v1' | '2v2'
    maxTeams: number
    currentTeams: number
    startDate: string
    endDate?: string
    location: string
    isPublic: boolean
    createdAt: string
    updatedAt: string
}

export interface LeaguePlayer {
    leagueId: string
    playerId: string
    playerName: string
    joinedAt: string
    role: 'player' | 'organizer'
}

export interface CreateLeagueData {
    name: string
    description: string
    format: '1v1' | '2v2'
    maxTeams: number
    startDate: string
    endDate?: string
    location: string
    isPublic: boolean
}