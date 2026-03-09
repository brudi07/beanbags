<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { usePlayerStore } from '~/stores/playerStore'
import type { GameFormat } from '~/types/game'

const router = useRouter()
const playerStore = usePlayerStore()

const gameFormat = ref<GameFormat>('2v2')
const team1Player1 = ref('')
const team1Player2 = ref('')
const team2Player1 = ref('')
const team2Player2 = ref('')

const canStartGame = computed(() => {
  if (gameFormat.value === '1v1') {
    return team1Player1.value.trim() && team2Player1.value.trim()
  } else {
    return team1Player1.value.trim() && team1Player2.value.trim() &&
      team2Player1.value.trim() && team2Player2.value.trim()
  }
})

function startGame() {
  const playerNames = {
    team1: gameFormat.value === '1v1'
      ? [team1Player1.value.trim()]
      : [team1Player1.value.trim(), team1Player2.value.trim()],
    team2: gameFormat.value === '1v1'
      ? [team2Player1.value.trim()]
      : [team2Player1.value.trim(), team2Player2.value.trim()]
  }

  playerStore.setupGame(gameFormat.value, playerNames)

  // Navigate to the game detail page - update ID as needed
  // For now using ID 1, but you'll want to generate/use actual game ID
  router.push('/games/1')
}
</script>

<template>
  <div class="max-w-2xl mx-auto p-6 space-y-6">

    <div class="text-center">
      <h1 class="text-3xl font-bold mb-2">Setup Game</h1>
      <p class="text-gray-600">Choose format and enter player names</p>
    </div>

    <!-- Game Format Selection -->
    <div class="bg-white rounded-xl border shadow p-6">
      <h2 class="text-lg font-semibold mb-4">Game Format</h2>

      <div class="flex gap-4">
        <button @click="gameFormat = '1v1'" class="flex-1 py-3 px-6 rounded-lg font-semibold transition-all" :class="gameFormat === '1v1'
          ? 'bg-blue-600 text-white'
          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'">
          1v1
        </button>

        <button @click="gameFormat = '2v2'" class="flex-1 py-3 px-6 rounded-lg font-semibold transition-all" :class="gameFormat === '2v2'
          ? 'bg-blue-600 text-white'
          : 'bg-gray-100 text-gray-700 hover:bg-gray-200'">
          2v2
        </button>
      </div>
    </div>

    <!-- Player Names -->
    <div class="grid grid-cols-2 gap-6">

      <!-- Team 1 -->
      <div class="bg-red-50 rounded-xl border-2 border-red-200 p-6">
        <h2 class="text-lg font-semibold mb-4 text-red-700">Team 1</h2>

        <div class="space-y-3">
          <input v-model="team1Player1" type="text" placeholder="Player 1 Name"
            class="w-full px-4 py-2 rounded-lg border border-red-300 focus:outline-none focus:ring-2 focus:ring-red-500" />

          <input v-if="gameFormat === '2v2'" v-model="team1Player2" type="text" placeholder="Player 2 Name"
            class="w-full px-4 py-2 rounded-lg border border-red-300 focus:outline-none focus:ring-2 focus:ring-red-500" />
        </div>
      </div>

      <!-- Team 2 -->
      <div class="bg-blue-50 rounded-xl border-2 border-blue-200 p-6">
        <h2 class="text-lg font-semibold mb-4 text-blue-700">Team 2</h2>

        <div class="space-y-3">
          <input v-model="team2Player1" type="text" placeholder="Player 1 Name"
            class="w-full px-4 py-2 rounded-lg border border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />

          <input v-if="gameFormat === '2v2'" v-model="team2Player2" type="text" placeholder="Player 2 Name"
            class="w-full px-4 py-2 rounded-lg border border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
      </div>

    </div>

    <!-- Start Button -->
    <div class="flex justify-center">
      <button @click="startGame" :disabled="!canStartGame"
        class="px-8 py-4 rounded-lg font-semibold text-lg transition-all" :class="canStartGame
          ? 'bg-green-600 text-white hover:bg-green-700 cursor-pointer'
          : 'bg-gray-300 text-gray-500 cursor-not-allowed'">
        Start Game
      </button>
    </div>

  </div>
</template>