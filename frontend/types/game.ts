export interface Game {
    id: string
    leagueId?: string
    team1: string
    team2: string
    status: "pending" | "active" | "completed" | "abandoned"
    createdAt: string
}

export interface ThrowData {
    team: 1 | 2
    player: string
    bagNumber: 1 | 2 | 3 | 4
    result: 'board' | 'hole' | 'miss'
}