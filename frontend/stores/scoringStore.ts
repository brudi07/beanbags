import { defineStore } from 'pinia'
import type { ThrowData } from '~/types/game'

export const useScoringStore = defineStore('scoring', {

    state: () => ({
        round: 1,
        bags: [] as ThrowData[],
        rounds: [] as { round: number; throws: ThrowData[] }[],
        team1Score: 0,
        team2Score: 0
    }),

    actions: {

        addThrow(throwData: ThrowData) {
            this.bags.push(throwData)
        },

        finishRound() {
            this.rounds.push({
                round: this.round,
                throws: this.bags
            })

            this.bags = []
            this.round++
        }

    }

})