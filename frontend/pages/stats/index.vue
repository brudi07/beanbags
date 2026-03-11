<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'

const api = useApi()

const stats = ref<any>(null)
const isLoading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
    try {
        stats.value = await api.fetch('/players/me/stats')
    } catch (err: any) {
        error.value = 'Failed to load stats'
    } finally {
        isLoading.value = false
    }
})
</script>

<template>
    <div class="max-w-6xl mx-auto p-6">

        <h1 class="text-3xl font-bold mb-8">My Stats</h1>

        <!-- Loading -->
        <div v-if="isLoading" class="text-gray-500">Loading stats...</div>

        <!-- Error -->
        <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">
            {{ error }}
        </div>

        <!-- No games yet -->
        <div v-else-if="!stats || stats.total_throws === 0"
            class="text-center py-16 text-gray-500">
            <p class="text-xl font-semibold mb-2">No games played yet</p>
            <p class="text-sm">Play some games and your stats will show up here.</p>
        </div>

        <!-- Stats grid -->
        <div v-else class="space-y-6">

            <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Total Throws</p>
                    <p class="text-3xl font-bold">{{ stats.total_throws }}</p>
                </div>
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Accuracy</p>
                    <p class="text-3xl font-bold">{{ stats.accuracy.toFixed(1) }}%</p>
                    <p class="text-xs text-gray-400 mt-1">Holes + boards</p>
                </div>
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Hole Shots</p>
                    <p class="text-3xl font-bold text-green-600">{{ stats.holes }}</p>
                </div>
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Points Contributed</p>
                    <p class="text-3xl font-bold text-blue-600">{{ stats.points_contributed }}</p>
                </div>
            </div>

            <div class="grid gap-4 sm:grid-cols-3">
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Board Hits</p>
                    <p class="text-2xl font-bold">{{ stats.boards }}</p>
                </div>
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">Misses</p>
                    <p class="text-2xl font-bold text-red-500">{{ stats.misses }}</p>
                </div>
                <div class="border rounded-xl p-6 bg-white shadow-sm">
                    <p class="text-sm text-gray-500 mb-1">ITOs</p>
                    <p class="text-2xl font-bold text-orange-500">{{ stats.itos }}</p>
                </div>
            </div>

        </div>

    </div>
</template>
