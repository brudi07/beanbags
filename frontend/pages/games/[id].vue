<script setup lang="ts">
import { computed } from 'vue'
import { useScoringStore } from '~/stores/scoringStore'
import ScoreBoard from '~/components/scoring/ScoreBoard.vue'
import CornholeBoard from '~/components/scoring/CornholeBoard.vue'
import RoundTracker from '~/components/scoring/RoundTracker.vue'

const scoringStore = useScoringStore()

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
  <div class="p-4 space-y-6">

    <ScoreBoard />

    <CornholeBoard />

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