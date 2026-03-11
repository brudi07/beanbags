<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'

const api = useApi()

interface GameEntry {
    id: number
    date: string
    team1: string
    team2: string
    winner: string | null
    user_team: number
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
        <div v-else class="space-y-3">
            <div v-for="game in games" :key="game.id"
                class="flex items-center justify-between border rounded-xl p-4 bg-white shadow-sm hover:bg-gray-50 transition-colors">

                <div>
                    <p class="font-semibold text-gray-900">{{ game.team1 }} vs {{ game.team2 }}</p>
                    <p class="text-sm text-gray-500">{{ formatDate(game.date) }}</p>
                </div>

                <span class="text-sm font-semibold px-3 py-1 rounded-full"
                    :class="resultLabel(game).cls">
                    {{ resultLabel(game).text }}
                </span>

            </div>
        </div>

    </div>
</template>
