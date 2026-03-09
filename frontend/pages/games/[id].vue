<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useScoringStore } from '~/stores/scoringStore'
import { usePlayerStore } from '~/stores/playerStore'
import ScoreBoard from '~/components/scoring/ScoreBoard.vue'
import CornholeBoard from '~/components/scoring/CornholeBoard.vue'
import RoundTracker from '~/components/scoring/RoundTracker.vue'
import PlayerStatsPanel from '~/components/scoring/PlayerStatsPanel.vue'

const route = useRoute()
const router = useRouter()
const scoringStore = useScoringStore()
const playerStore = usePlayerStore()

const isSubmitting = ref(false)
const submitError = ref<string | null>(null)

const allBagsPlaced = computed(() => {
  return scoringStore.team1BagsRemaining === 0 && scoringStore.team2BagsRemaining === 0
})

const buttonText = computed(() => {
  if (scoringStore.gameCompleted) {
    return 'Submit Final Score'
  }
  if (scoringStore.isViewingPastRound) {
    return 'Update Round'
  }
  if (!allBagsPlaced.value) {
    const placed = 8 - scoringStore.team1BagsRemaining - scoringStore.team2BagsRemaining
    return `Score Round (${placed}/8)`
  }
  return 'Score Round'
})

const buttonClass = computed(() => {
  if (scoringStore.gameCompleted && allBagsPlaced.value) {
    return 'bg-blue-600 text-white hover:bg-blue-700 cursor-pointer'
  }
  if (allBagsPlaced.value) {
    return 'bg-green-600 text-white hover:bg-green-700 cursor-pointer'
  }
  return 'bg-gray-300 text-gray-500 cursor-not-allowed'
})

async function handleButtonClick() {
  if (!allBagsPlaced.value) return

  if (scoringStore.gameCompleted) {
    // Submit final score
    await submitGameResults()
  } else {
    // Score the round
    scoringStore.scoreRound()
  }
}

async function submitGameResults() {
  isSubmitting.value = true
  submitError.value = null

  try {
    const gameId = route.params.id as string
    await scoringStore.submitGameResults(gameId)

    // Navigate to game results or games list
    router.push(`/games/${gameId}/results`) // Or wherever you want to go after submission
  } catch (error) {
    submitError.value = 'Failed to submit game results. Please try again.'
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <!-- Player Setup Screen -->
  <div v-if="!playerStore.gameStarted" class="min-h-screen flex items-center justify-center bg-gray-50">
    <PlayerSetup />
  </div>

  <!-- Game Screen -->
  <div v-else class="p-4 space-y-6">

    <ScoreBoard />

    <!-- Winner Banner -->
    <div v-if="scoringStore.gameCompleted"
      class="bg-gradient-to-r from-yellow-400 to-yellow-500 rounded-xl p-6 text-center shadow-lg">
      <h2 class="text-3xl font-bold text-gray-900 mb-2">
        🎉 Team {{ scoringStore.gameWinner }} Wins! 🎉
      </h2>
      <p class="text-lg text-gray-800">Final Score: {{ scoringStore.team1Score }} - {{ scoringStore.team2Score }}</p>
    </div>

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

    <!-- Score Round / Submit Button -->
    <div class="flex flex-col items-center gap-2 w-full">
      <button class="px-6 py-3 rounded-lg font-semibold transition-all" :class="buttonClass"
        :disabled="!allBagsPlaced || isSubmitting" @click="handleButtonClick">
        <span v-if="isSubmitting">Submitting...</span>
        <span v-else>{{ buttonText }}</span>
      </button>

      <p v-if="submitError" class="text-red-600 text-sm">{{ submitError }}</p>
    </div>

    <RoundTracker />

  </div>
</template>