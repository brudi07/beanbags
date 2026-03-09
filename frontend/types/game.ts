export interface Game {
    id: string
    leagueId?: string
    team1: string
    team2: string
    status: "pending" | "active" | "completed" | "abandoned"
    createdAt: string
}

export interface ThrowData {
    id: string
    round: number
    team: 1 | 2
    playerId: string
    x: number
    y: number
    result: 'hole' | 'board' | 'miss' | 'ito' // intentional throw off
    rotation: number
    timestamp: number
}

export interface RoundResult {
    round: number
    throws: ThrowData[]
    team1Points: number
    team2Points: number
}