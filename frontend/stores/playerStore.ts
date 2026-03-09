import { defineStore } from 'pinia'
import type { Player, PlayerStats, GameFormat } from '~/types/game'

export const usePlayerStore = defineStore('player', {
    state: () => ({
        players: [] as Player[],
        gameFormat: '2v2' as GameFormat,
        currentTeam1PlayerIndex: 0,
        currentTeam2PlayerIndex: 0,
        gameStarted: false
    }),

    getters: {
        team1Players(): Player[] {
            return this.players.filter(p => p.team === 1)
        },

        team2Players(): Player[] {
            return this.players.filter(p => p.team === 2)
        },

        currentTeam1Player(): Player | null {
            const team1 = this.team1Players
            return team1[this.currentTeam1PlayerIndex] || null
        },

        currentTeam2Player(): Player | null {
            const team2 = this.team2Players
            return team2[this.currentTeam2PlayerIndex] || null
        },

        getPlayerById(): (id: string) => Player | undefined {
            return (id: string) => this.players.find(p => p.id === id)
        }
    },

    actions: {
        setupGame(format: GameFormat, playerNames: { team1: string[], team2: string[] }) {
            this.gameFormat = format
            this.players = []

            // Create Team 1 players
            playerNames.team1.forEach((name, index) => {
                this.players.push({
                    id: `team1-player${index + 1}`,
                    name,
                    team: 1
                })
            })

            // Create Team 2 players
            playerNames.team2.forEach((name, index) => {
                this.players.push({
                    id: `team2-player${index + 1}`,
                    name,
                    team: 2
                })
            })

            // Start with first player from each team
            this.currentTeam1PlayerIndex = 0
            this.currentTeam2PlayerIndex = 0

            this.gameStarted = true
        },

        advanceTurn() {
            // In 2v2, advance to next player on each team
            if (this.gameFormat === '2v2') {
                const team1Count = this.team1Players.length
                const team2Count = this.team2Players.length

                if (team1Count > 0) {
                    this.currentTeam1PlayerIndex = (this.currentTeam1PlayerIndex + 1) % team1Count
                }
                if (team2Count > 0) {
                    this.currentTeam2PlayerIndex = (this.currentTeam2PlayerIndex + 1) % team2Count
                }
            }
            // In 1v1, players don't change
        },

        setTurnToPlayer(playerId: string) {
            const player = this.players.find(p => p.id === playerId)
            if (!player) return

            if (player.team === 1) {
                const index = this.team1Players.findIndex(p => p.id === playerId)
                if (index >= 0) this.currentTeam1PlayerIndex = index
            } else {
                const index = this.team2Players.findIndex(p => p.id === playerId)
                if (index >= 0) this.currentTeam2PlayerIndex = index
            }
        },

        resetGame() {
            this.players = []
            this.currentTeam1PlayerIndex = 0
            this.currentTeam2PlayerIndex = 0
            this.gameStarted = false
        },

        calculatePlayerStats(throws: any[], roundHistory?: any[]): PlayerStats[] {
            const statsMap = new Map<string, PlayerStats>()

            // Initialize stats for all players
            this.players.forEach(player => {
                statsMap.set(player.id, {
                    playerId: player.id,
                    holes: 0,
                    boards: 0,
                    misses: 0,
                    itos: 0,
                    totalThrows: 0,
                    pointsContributed: 0,
                    accuracy: 0,
                    pointsPerRound: 0,
                    differentialPerRound: 0,
                    busts: 0
                })
            })

            // Track points by round for each player
            const roundPointsMap = new Map<string, number[]>() // playerId -> array of points per round

            this.players.forEach(player => {
                roundPointsMap.set(player.id, [])
            })

            // Calculate stats from all throws (current + history)
            const allThrows = [...throws]

            allThrows.forEach(throwData => {
                const stats = statsMap.get(throwData.playerId)
                if (!stats) return

                stats.totalThrows++

                switch (throwData.result) {
                    case 'hole':
                        stats.holes++
                        stats.pointsContributed += 3
                        break
                    case 'board':
                        stats.boards++
                        stats.pointsContributed += 1
                        break
                    case 'miss':
                        stats.misses++
                        break
                    case 'ito':
                        stats.itos++
                        break
                }

                // Calculate accuracy (holes + boards / total throws)
                const scoringThrows = stats.holes + stats.boards
                stats.accuracy = stats.totalThrows > 0
                    ? Math.round((scoringThrows / stats.totalThrows) * 100)
                    : 0
            })

            // Calculate per-round stats from round history
            if (roundHistory && roundHistory.length > 0) {
                roundHistory.forEach(round => {
                    // Track points each player contributed in this round
                    const roundPlayerPoints = new Map<string, number>()

                    round.throws.forEach((throwData: any) => {
                        const currentPoints = roundPlayerPoints.get(throwData.playerId) || 0
                        let pointsToAdd = 0

                        if (throwData.result === 'hole') pointsToAdd = 3
                        else if (throwData.result === 'board') pointsToAdd = 1

                        roundPlayerPoints.set(throwData.playerId, currentPoints + pointsToAdd)
                    })

                    // Check for busts and attribute to players who threw in that round
                    if (round.team1Busted) {
                        // Find team 1 players who threw this round
                        const team1ThrowersThisRound = new Set<string>()
                        round.throws.forEach((throwData: any) => {
                            const player = this.players.find(p => p.id === throwData.playerId)
                            if (player?.team === 1) {
                                team1ThrowersThisRound.add(throwData.playerId)
                            }
                        })
                        // Each team 1 player who threw gets credited with the bust
                        team1ThrowersThisRound.forEach(playerId => {
                            const stats = statsMap.get(playerId)
                            if (stats) stats.busts++
                        })
                    }

                    if (round.team2Busted) {
                        // Find team 2 players who threw this round
                        const team2ThrowersThisRound = new Set<string>()
                        round.throws.forEach((throwData: any) => {
                            const player = this.players.find(p => p.id === throwData.playerId)
                            if (player?.team === 2) {
                                team2ThrowersThisRound.add(throwData.playerId)
                            }
                        })
                        // Each team 2 player who threw gets credited with the bust
                        team2ThrowersThisRound.forEach(playerId => {
                            const stats = statsMap.get(playerId)
                            if (stats) stats.busts++
                        })
                    }

                    // Add round points to each player's round array
                    this.players.forEach(player => {
                        const pointsThisRound = roundPlayerPoints.get(player.id) || 0
                        const playerRounds = roundPointsMap.get(player.id)
                        if (playerRounds) {
                            playerRounds.push(pointsThisRound)
                        }
                    })
                })

                // Calculate averages
                statsMap.forEach((stats, playerId) => {
                    const playerRounds = roundPointsMap.get(playerId) || []
                    const roundsPlayed = playerRounds.length

                    if (roundsPlayed > 0) {
                        // Points per round
                        const totalPoints = playerRounds.reduce((sum, pts) => sum + pts, 0)
                        stats.pointsPerRound = Math.round((totalPoints / roundsPlayed) * 10) / 10

                        // Differential per round (team's scoring vs opponent's scoring)
                        const player = this.players.find(p => p.id === playerId)
                        if (player) {
                            let totalDifferential = 0

                            roundHistory.forEach(round => {
                                if (player.team === 1) {
                                    totalDifferential += round.team1Points
                                } else {
                                    totalDifferential += round.team2Points
                                }
                            })

                            stats.differentialPerRound = Math.round((totalDifferential / roundsPlayed) * 10) / 10
                        }
                    }
                })
            }

            return Array.from(statsMap.values())
        }
    }
})