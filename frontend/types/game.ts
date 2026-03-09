export interface Game {
    id: string
    leagueId?: string
    team1: string
    team2: string
    status: "pending" | "active" | "completed" | "abandoned"
    createdAt: string
}

export interface Player {
    id: string
    name: string
    team: 1 | 2
}

export interface PlayerStats {
    playerId: string
    holes: number
    boards: number
    misses: number
    itos: number
    totalThrows: number
    pointsContributed: number
    accuracy: number // percentage of scoring attempts (excludes ITOs)
    pointsPerRound: number // average points contributed per round
    differentialPerRound: number // average differential impact per round
    busts: number // number of times player caused their team to bust
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
    team1Busted?: boolean
    team2Busted?: boolean
}

export type GameFormat = '1v1' | '2v2'