<script setup lang="ts">
import { computed } from 'vue'
import { usePlayerStore } from '~/stores/playerStore'
import { useScoringStore } from '~/stores/scoringStore'
import type { Player, PlayerStats } from '~/types/game'

const props = defineProps<{
    team: 1 | 2
}>()

const playerStore = usePlayerStore()
const scoringStore = useScoringStore()

const teamPlayers = computed(() =>
    props.team === 1 ? playerStore.team1Players : playerStore.team2Players
)

const playerStats = computed(() => {
    // Collect all throws from history + current round
    const allThrows: any[] = []

    // Add throws from completed rounds
    scoringStore.roundHistory.forEach(round => {
        allThrows.push(...round.throws)
    })

    // Add current round throws
    allThrows.push(...scoringStore.throws)

    const allStats = playerStore.calculatePlayerStats(allThrows, scoringStore.roundHistory)
    return teamPlayers.value.map(player => {
        const stats = allStats.find(s => s.playerId === player.id)
        return {
            player,
            stats: stats || {
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
            }
        }
    })
})

const isCurrentTurn = (playerId: string) => {
    if (props.team === 1) {
        return playerStore.currentTeam1Player?.id === playerId
    } else {
        return playerStore.currentTeam2Player?.id === playerId
    }
}
</script>

<template>
    <div class="space-y-4">

        <div v-for="{ player, stats } in playerStats" :key="player.id"
            class="bg-white rounded-xl border shadow-sm p-4 transition-all" :class="[
                isCurrentTurn(player.id) ? 'ring-2 ring-offset-2' : '',
                team === 1 ? (isCurrentTurn(player.id) ? 'ring-red-500' : '') : (isCurrentTurn(player.id) ? 'ring-blue-500' : '')
            ]">

            <!-- Player Name -->
            <div class="flex items-center gap-2 mb-3">
                <div class="w-3 h-3 rounded-full" :class="team === 1 ? 'bg-red-500' : 'bg-blue-500'" />
                <h3 class="font-semibold text-lg">{{ player.name }}</h3>
                <span v-if="isCurrentTurn(player.id)" class="ml-auto text-xs font-semibold px-2 py-1 rounded"
                    :class="team === 1 ? 'bg-red-100 text-red-700' : 'bg-blue-100 text-blue-700'">
                    Current
                </span>
            </div>

            <!-- Stats Grid -->
            <div class="grid grid-cols-2 gap-3 text-sm">

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Points</div>
                    <div class="font-bold text-lg">{{ stats.pointsContributed }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Accuracy</div>
                    <div class="font-bold text-lg">{{ stats.accuracy }}%</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Pts/Round</div>
                    <div class="font-bold">{{ stats.pointsPerRound }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Diff/Round</div>
                    <div class="font-bold">{{ stats.differentialPerRound }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Holes</div>
                    <div class="font-bold">{{ stats.holes }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Boards</div>
                    <div class="font-bold">{{ stats.boards }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">Off</div>
                    <div class="font-bold text-gray-600">{{ stats.misses }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2">
                    <div class="text-gray-500 text-xs">ITOs</div>
                    <div class="font-bold text-gray-600">{{ stats.itos }}</div>
                </div>

                <div class="bg-gray-50 rounded-lg p-2 col-span-2">
                    <div class="text-gray-500 text-xs">Busts</div>
                    <div class="font-bold text-red-600 text-lg">{{ stats.busts }}</div>
                </div>

            </div>

        </div>

    </div>
</template>