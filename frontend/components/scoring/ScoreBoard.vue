<script setup lang="ts">
import { useScoringStore } from '~/stores/scoringStore'

const scoringStore = useScoringStore()
</script>

<template>
    <div class="border rounded-xl p-4 bg-white shadow">

        <div class="flex items-center justify-between gap-4">

            <!-- Previous Round Arrow -->
            <button @click="scoringStore.goToPreviousRound()" :disabled="!scoringStore.canGoBack"
                class="p-2 rounded-lg transition-all" :class="scoringStore.canGoBack
                    ? 'hover:bg-gray-100 text-gray-700 cursor-pointer'
                    : 'text-gray-300 cursor-not-allowed'">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
                </svg>
            </button>

            <!-- Team 1 Score -->
            <div class="text-center flex-1">
                <div class="relative inline-block">
                    <!-- Crown for honors -->
                    <div v-if="scoringStore.teamWithHonors === 1"
                        class="absolute -top-8 left-1/2 -translate-x-1/2 text-3xl">
                        👑
                    </div>
                    <p class="text-sm text-red-500">Team 1</p>
                </div>
                <p class="text-6xl font-bold text-red-500">
                    {{ scoringStore.displayedTeam1Score }}
                </p>
            </div>

            <!-- Round Display -->
            <div class="text-center px-6">
                <p class="text-sm text-gray-500">Round</p>
                <p class="text-5xl font-bold"
                    :class="scoringStore.isViewingPastRound ? 'text-orange-600' : 'text-gray-900'">
                    {{ scoringStore.currentRoundView }}
                </p>
                <p v-if="scoringStore.isViewingPastRound" class="text-xs text-orange-600 font-semibold mt-1">
                    Editing
                </p>
            </div>

            <!-- Team 2 Score -->
            <div class="text-center flex-1">
                <div class="relative inline-block">
                    <!-- Crown for honors -->
                    <div v-if="scoringStore.teamWithHonors === 2"
                        class="absolute -top-8 left-1/2 -translate-x-1/2 text-3xl">
                        👑
                    </div>
                    <p class="text-sm text-blue-500">Team 2</p>
                </div>
                <p class="text-6xl font-bold text-blue-500">
                    {{ scoringStore.displayedTeam2Score }}
                </p>
            </div>

            <!-- Next Round Arrow -->
            <button @click="scoringStore.goToNextRound()" :disabled="!scoringStore.canGoForward"
                class="p-2 rounded-lg transition-all" :class="scoringStore.canGoForward
                    ? 'hover:bg-gray-100 text-gray-700 cursor-pointer'
                    : 'text-gray-300 cursor-not-allowed'">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24"
                    stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
                </svg>
            </button>

        </div>

    </div>
</template>