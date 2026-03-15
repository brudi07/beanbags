<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'

const api = useApi()

interface PlayerStat {
    player_id: number
    player_name: string
    team: number
    holes: number
    boards: number
    misses: number
    itos: number
    four_bagger: number
    is_me: boolean
}

interface GameEntry {
    id: number
    date: string
    team1: string
    team2: string
    winner: string | null
    user_team: number
    team1_score: number
    team2_score: number
    player_stats: PlayerStat[]
}

const games = ref<GameEntry[]>([])
const isLoading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
    try {
        games.value = await api.fetch<GameEntry[]>('/players/me/games')
    } catch (err: any) {
        error.value = 'Failed to load game history'
    } finally {
        isLoading.value = false
    }
})

function formatDate(dateStr: string) {
    const d = new Date(dateStr)
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
}

function resultLabel(game: GameEntry) {
    if (!game.winner) return { text: 'In Progress', cls: 'text-gray-500 bg-gray-100' }
    const userTeamName = game.user_team === 1 ? game.team1 : game.team2
    const won = game.winner === userTeamName
    return won
        ? { text: 'Win', cls: 'text-green-700 bg-green-100' }
        : { text: 'Loss', cls: 'text-red-700 bg-red-100' }
}

function accuracy(stat: PlayerStat) {
    const attempts = stat.holes + stat.boards + stat.misses
    if (attempts === 0) return 0
    return Math.round((stat.holes + stat.boards) / attempts * 100)
}

function teamPlayers(game: GameEntry, team: number) {
    return game.player_stats.filter(p => p.team === team)
}
</script>

<template>
    <div class="max-w-5xl mx-auto p-6">

        <h1 class="text-3xl font-bold mb-8">Game History</h1>

        <!-- Loading -->
        <div v-if="isLoading" class="text-gray-500">Loading history...</div>

        <!-- Error -->
        <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
            {{ error }}
        </div>

        <!-- Empty state -->
        <div v-else-if="games.length === 0" class="text-center py-16 text-gray-500">
            <p class="text-xl font-semibold mb-2">No games played yet</p>
            <p class="text-sm">Your completed games will appear here.</p>
        </div>

        <!-- Game list -->
        <div v-else class="space-y-4">
            <div v-for="game in games" :key="game.id"
                class="border rounded-xl p-5 bg-white shadow-sm">

                <!-- Top row: matchup, score, result badge -->
                <div class="flex items-start justify-between gap-4">
                    <div>
                        <p class="font-semibold text-gray-900 text-lg">
                            {{ game.team1 }} vs {{ game.team2 }}
                        </p>
                        <p class="text-sm text-gray-500 mt-0.5">{{ formatDate(game.date) }}</p>
                    </div>

                    <div class="flex items-center gap-3 shrink-0">
                        <!-- Score -->
                        <div v-if="game.winner" class="text-center">
                            <p class="text-2xl font-bold tabular-nums leading-none">
                                <span :class="game.winner === game.team1 ? 'text-green-600' : 'text-gray-800'">
                                    {{ game.team1_score }}
                                </span>
                                <span class="text-gray-400 mx-1">–</span>
                                <span :class="game.winner === game.team2 ? 'text-green-600' : 'text-gray-800'">
                                    {{ game.team2_score }}
                                </span>
                            </p>
                        </div>

                        <!-- Win/Loss badge -->
                        <span class="text-sm font-semibold px-3 py-1 rounded-full"
                            :class="resultLabel(game).cls">
                            {{ resultLabel(game).text }}
                        </span>
                    </div>
                </div>

                <!-- Player stats -->
                <div v-if="game.player_stats.length > 0"
                    class="mt-4 pt-4 border-t border-gray-100 space-y-3">

                    <template v-for="team in [1, 2]" :key="team">
                        <div v-if="teamPlayers(game, team).length > 0">
                            <!-- Team label -->
                            <p class="text-xs font-semibold text-gray-400 uppercase tracking-wide mb-1">
                                {{ team === 1 ? game.team1 : game.team2 }}
                            </p>

                            <!-- Per-player rows -->
                            <div v-for="stat in teamPlayers(game, team)" :key="stat.player_id"
                                class="flex items-center gap-3 py-1.5">
                                <!-- Player name -->
                                <div class="w-32 flex items-center gap-1.5 shrink-0">
                                    <span class="text-sm font-medium text-gray-800 truncate">{{ stat.player_name }}</span>
                                    <span v-if="stat.is_me"
                                        class="text-xs font-semibold bg-blue-100 text-blue-700 px-1.5 py-0.5 rounded shrink-0">You</span>
                                </div>

                                <!-- Stats -->
                                <div class="grid gap-2 text-center flex-1"
                                    :class="stat.four_bagger > 0 ? 'grid-cols-6' : 'grid-cols-5'">
                                    <div>
                                        <p class="text-sm font-bold text-blue-600">{{ stat.holes }}</p>
                                        <p class="text-xs text-gray-400">Holes</p>
                                    </div>
                                    <div>
                                        <p class="text-sm font-bold text-green-600">{{ stat.boards }}</p>
                                        <p class="text-xs text-gray-400">Boards</p>
                                    </div>
                                    <div>
                                        <p class="text-sm font-bold text-red-500">{{ stat.misses }}</p>
                                        <p class="text-xs text-gray-400">Misses</p>
                                    </div>
                                    <div>
                                        <p class="text-sm font-bold text-orange-500">{{ stat.itos }}</p>
                                        <p class="text-xs text-gray-400">ITOs</p>
                                    </div>
                                    <div>
                                        <p class="text-sm font-bold text-purple-600">{{ accuracy(stat) }}%</p>
                                        <p class="text-xs text-gray-400">Accuracy</p>
                                    </div>
                                    <div v-if="stat.four_bagger > 0">
                                        <p class="text-sm font-bold text-yellow-500">{{ stat.four_bagger }}</p>
                                        <p class="text-xs text-gray-400">4-Bagger</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </template>
                </div>

            </div>
        </div>

    </div>
</template>
