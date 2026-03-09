<script setup lang="ts">
import { computed } from 'vue'
import { useScoringStore } from '~/stores/scoringStore'
import { usePlayerStore } from '~/stores/playerStore'
import ScoreBoard from '~/components/scoring/ScoreBoard.vue'
import CornholeBoard from '~/components/scoring/CornholeBoard.vue'
import RoundTracker from '~/components/scoring/RoundTracker.vue'
import PlayerStatsPanel from '~/components/scoring/PlayerStatsPanel.vue'

const scoringStore = useScoringStore()
const playerStore = usePlayerStore()

const allBagsPlaced = computed(() => {
  return scoringStore.team1BagsRemaining === 0 && scoringStore.team2BagsRemaining === 0
})

const buttonText = computed(() => {
  if (scoringStore.isViewingPastRound) {
    return 'Update Round'
  }
  if (!allBagsPlaced.value) {
    const placed = 8 - scoringStore.team1BagsRemaining - scoringStore.team2BagsRemaining
    return `Score Round (${placed}/8)`
  }
  return 'Score Round'
})
</script>

<template>
  <!-- Player Setup Screen -->
  <div v-if="!playerStore.gameStarted" class="min-h-screen flex items-center justify-center bg-gray-50">
    <PlayerSetup />
  </div>

  <!-- Game Screen -->
  <div v-else class="p-4 space-y-6">

    <ScoreBoard />

    <!-- Main Game Area with Stats Panels -->
    <div class="flex gap-6 justify-center items-start">

      <!-- Team 1 Stats (Left) -->
      <div class="hidden lg:block w-64">
        <PlayerStatsPanel :team="1" />
      </div>

      <!-- Cornhole Board (Center) -->
      <div class="flex-shrink-0">
        <CornholeBoard />
      </div>

      <!-- Team 2 Stats (Right) -->
      <div class="hidden lg:block w-64">
        <PlayerStatsPanel :team="2" />
      </div>

    </div>

    <!-- Mobile Stats (Below board on smaller screens) -->
    <div class="lg:hidden grid grid-cols-2 gap-4">
      <PlayerStatsPanel :team="1" />
      <PlayerStatsPanel :team="2" />
    </div>

    <!-- Score Round Button -->
    <div class="flex justify-center w-full">
      <button class="px-6 py-3 rounded-lg font-semibold transition-all" :class="allBagsPlaced
        ? 'bg-green-600 text-white hover:bg-green-700 cursor-pointer'
        : 'bg-gray-300 text-gray-500 cursor-not-allowed'" :disabled="!allBagsPlaced"
        @click="scoringStore.scoreRound()">
        {{ buttonText }}
      </button>
    </div>

    <RoundTracker />

  </div>
</template>