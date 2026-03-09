import { defineStore } from 'pinia'
import type { ThrowData, RoundResult } from '~/types/game'
import { usePlayerStore } from './playerStore'

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
        teamWithHonors: null as 1 | 2 | null
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

        addThrow(throwData: ThrowData) {
            this.throws.push(throwData)
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

                this.round++
                this.currentRoundView = this.round

                // Advance to next players' turn (2v2: switch to other players on each team)
                const playerStore = usePlayerStore()
                playerStore.advanceTurn()
            }

            this.resetRound()
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
        }

    }
})