import { defineStore } from 'pinia'
import type { ThrowData, RoundResult } from '~/types/game'
import { usePlayerStore } from './playerStore'
import { useApi } from '~/composables/useApi'

const STORAGE_KEY = 'beanbags_active_game'

export const useScoringStore = defineStore('scoring', {
    state: () => ({
        team1Score: 0,
        team2Score: 0,
        round: 1,

        throws: [] as ThrowData[],

        team1BagsRemaining: 4,
        team2BagsRemaining: 4,

        roundHistory: [] as RoundResult[],

        // Navigation state
        currentRoundView: 1,
        isViewingPastRound: false,

        // Honors tracking (which team scored last)
        teamWithHonors: null as 1 | 2 | null,

        // Game completion
        gameWinner: null as 1 | 2 | null,
        gameCompleted: false,

        // Persistence tracking
        activeGameId: null as string | null,
        activeLeagueId: null as string | null,
    }),

    getters: {
        canGoBack(): boolean {
            return this.currentRoundView > 1
        },

        canGoForward(): boolean {
            return this.currentRoundView < this.round
        },

        displayedTeam1Score(): number {
            if (!this.isViewingPastRound) return this.team1Score

            // Calculate score up to current viewed round
            let score = 0
            for (let i = 0; i < this.currentRoundView; i++) {
                const round = this.roundHistory[i]
                if (round) {
                    score += round.team1Points
                }
            }
            return score
        },

        displayedTeam2Score(): number {
            if (!this.isViewingPastRound) return this.team2Score

            // Calculate score up to current viewed round
            let score = 0
            for (let i = 0; i < this.currentRoundView; i++) {
                const round = this.roundHistory[i]
                if (round) {
                    score += round.team2Points
                }
            }
            return score
        }
    },

    actions: {

        setActiveGame(gameId: string, leagueId?: string) {
            // Reset all scoring state for a fresh game
            this.team1Score = 0
            this.team2Score = 0
            this.round = 1
            this.throws = []
            this.team1BagsRemaining = 4
            this.team2BagsRemaining = 4
            this.roundHistory = []
            this.currentRoundView = 1
            this.isViewingPastRound = false
            this.teamWithHonors = null
            this.gameWinner = null
            this.gameCompleted = false
            // Set new game context
            this.activeGameId = gameId
            this.activeLeagueId = leagueId ?? null
            this._saveToStorage()
        },

        _saveToStorage() {
            if (!this.activeGameId || this.activeGameId === 'pickup') return
            const playerStore = usePlayerStore()
            try {
                localStorage.setItem(STORAGE_KEY, JSON.stringify({
                    gameId: this.activeGameId,
                    leagueId: this.activeLeagueId,
                    team1Score: this.team1Score,
                    team2Score: this.team2Score,
                    round: this.round,
                    throws: this.throws,
                    team1BagsRemaining: this.team1BagsRemaining,
                    team2BagsRemaining: this.team2BagsRemaining,
                    roundHistory: this.roundHistory,
                    currentRoundView: this.currentRoundView,
                    isViewingPastRound: this.isViewingPastRound,
                    teamWithHonors: this.teamWithHonors,
                    gameWinner: this.gameWinner,
                    gameCompleted: this.gameCompleted,
                    players: playerStore.players,
                    gameFormat: playerStore.gameFormat,
                    currentTeam1PlayerIndex: playerStore.currentTeam1PlayerIndex,
                    currentTeam2PlayerIndex: playerStore.currentTeam2PlayerIndex,
                }))
            } catch {
                // localStorage unavailable (private browsing quota, etc.) — silently continue
            }
        },

        restoreGameState(gameId: string): boolean {
            try {
                const raw = localStorage.getItem(STORAGE_KEY)
                if (!raw) return false
                const saved = JSON.parse(raw)
                if (saved.gameId !== gameId) return false

                this.activeGameId = saved.gameId
                this.activeLeagueId = saved.leagueId ?? null
                this.team1Score = saved.team1Score
                this.team2Score = saved.team2Score
                this.round = saved.round
                this.throws = saved.throws ?? []
                this.team1BagsRemaining = saved.team1BagsRemaining
                this.team2BagsRemaining = saved.team2BagsRemaining
                this.roundHistory = saved.roundHistory ?? []
                this.currentRoundView = saved.currentRoundView
                this.isViewingPastRound = saved.isViewingPastRound ?? false
                this.teamWithHonors = saved.teamWithHonors ?? null
                this.gameWinner = saved.gameWinner ?? null
                this.gameCompleted = saved.gameCompleted ?? false

                const playerStore = usePlayerStore()
                playerStore.players = saved.players ?? []
                playerStore.gameFormat = saved.gameFormat ?? '2v2'
                playerStore.currentTeam1PlayerIndex = saved.currentTeam1PlayerIndex ?? 0
                playerStore.currentTeam2PlayerIndex = saved.currentTeam2PlayerIndex ?? 0
                playerStore.gameStarted = true

                return true
            } catch {
                return false
            }
        },

        clearGameState() {
            this.activeGameId = null
            this.activeLeagueId = null
            try {
                localStorage.removeItem(STORAGE_KEY)
            } catch {
                // ignore
            }
        },

        addThrow(throwData: ThrowData) {
            this.throws.push(throwData)
            this._saveToStorage()
        },

        resetRound() {
            this.throws = []
            this.team1BagsRemaining = 4
            this.team2BagsRemaining = 4
        },

        scoreRound() {
            let team1 = 0
            let team2 = 0

            for (const throwData of this.throws) {

                if (throwData.result === 'hole') {
                    if (throwData.team === 1) team1 += 3
                    else team2 += 3
                }

                if (throwData.result === 'board') {
                    if (throwData.team === 1) team1 += 1
                    else team2 += 1
                }

            }

            // cancellation scoring
            let team1Points = 0
            let team2Points = 0

            if (team1 > team2) {
                team1Points = team1 - team2
                this.teamWithHonors = 1
            } else if (team2 > team1) {
                team2Points = team2 - team1
                this.teamWithHonors = 2
            } else {
                // No score this round, honors stays with whoever had it
            }

            // If editing a past round, update history and recalculate scores
            if (this.isViewingPastRound) {
                this.roundHistory[this.currentRoundView - 1] = {
                    round: this.currentRoundView,
                    throws: [...this.throws],
                    team1Points,
                    team2Points
                }
                this.recalculateScores()
                this.goToCurrentRound()
            } else {
                // Normal scoring for current round

                // Add points and check for bust (over 21)
                const newTeam1Score = this.team1Score + team1Points
                const newTeam2Score = this.team2Score + team2Points

                const team1Busted = newTeam1Score > 21
                const team2Busted = newTeam2Score > 21

                this.roundHistory.push({
                    round: this.round,
                    throws: [...this.throws],
                    team1Points,
                    team2Points,
                    team1Busted,
                    team2Busted
                })

                // Handle Team 1 bust
                if (team1Busted) {
                    this.team1Score = 15
                } else {
                    this.team1Score = newTeam1Score
                }

                // Handle Team 2 bust
                if (team2Busted) {
                    this.team2Score = 15
                } else {
                    this.team2Score = newTeam2Score
                }

                // Check for winner (exactly 21 points)
                if (this.team1Score === 21) {
                    this.gameWinner = 1
                    this.gameCompleted = true
                } else if (this.team2Score === 21) {
                    this.gameWinner = 2
                    this.gameCompleted = true
                }

                this.round++
                this.currentRoundView = this.round

                // Advance to next players' turn (2v2: switch to other players on each team)
                const playerStore = usePlayerStore()
                playerStore.advanceTurn()
            }

            this.resetRound()
            this._saveToStorage()
        },

        recalculateScores() {
            this.team1Score = 0
            this.team2Score = 0

            for (const roundResult of this.roundHistory) {
                // Add points and check for bust
                const newTeam1Score = this.team1Score + roundResult.team1Points
                const newTeam2Score = this.team2Score + roundResult.team2Points

                // Handle Team 1 bust
                if (newTeam1Score > 21) {
                    this.team1Score = 15
                } else {
                    this.team1Score = newTeam1Score
                }

                // Handle Team 2 bust
                if (newTeam2Score > 21) {
                    this.team2Score = 15
                } else {
                    this.team2Score = newTeam2Score
                }
            }

            // Recalculate honors from last completed round
            const lastRound = this.roundHistory[this.roundHistory.length - 1]
            if (lastRound?.team1Points && lastRound.team1Points > 0) {
                this.teamWithHonors = 1
            } else if (lastRound?.team2Points && lastRound.team2Points > 0) {
                this.teamWithHonors = 2
            }
        },

        goToPreviousRound() {
            if (this.canGoBack) {
                this.currentRoundView--
                this.loadRound(this.currentRoundView)
            }
        },

        goToNextRound() {
            if (this.canGoForward) {
                this.currentRoundView++
                if (this.currentRoundView === this.round) {
                    this.goToCurrentRound()
                } else {
                    this.loadRound(this.currentRoundView)
                }
            }
        },

        loadRound(roundNumber: number) {
            const roundData = this.roundHistory[roundNumber - 1]
            if (roundData) {
                this.isViewingPastRound = true
                this.throws = [...roundData.throws]
                this.team1BagsRemaining = 4 - this.throws.filter(t => t.team === 1).length
                this.team2BagsRemaining = 4 - this.throws.filter(t => t.team === 2).length
            }
        },

        goToCurrentRound() {
            this.isViewingPastRound = false
            this.currentRoundView = this.round
            this.resetRound()
        },

        async submitGameResults(gameId: string) {
            const api = useApi()

            const gameData = {
                winner: this.gameWinner,
                final_score: {
                    team1: this.team1Score,
                    team2: this.team2Score
                },
                total_rounds: this.round - 1,
                round_history: this.roundHistory.map(r => ({
                    round: r.round,
                    throws: [],
                    team1_points: r.team1Points,
                    team2_points: r.team2Points,
                    team1_busted: r.team1Busted ?? false,
                    team2_busted: r.team2Busted ?? false,
                })),
                completed_at: new Date().toISOString()
            }

            const result = await api.fetch(`/games/${gameId}/complete`, {
                method: 'POST',
                body: gameData
            })

            this.clearGameState()
            return result
        }

    }
})
