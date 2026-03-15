<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { usePlayerStore } from '~/stores/playerStore'
import { useApi } from '~/composables/useApi'
import { useToast } from '~/composables/useToast'
import type { GameFormat } from '~/types/game'

const router = useRouter()
const route = useRoute()
const api = useApi()
const toast = useToast()
const playerStore = usePlayerStore()

const leagueId = route.query.leagueId as string | undefined
const leagueGameId = route.query.leagueGameId as string | undefined
const isLeagueGame = !!(leagueId && leagueGameId)

const gameFormat = ref<GameFormat>('2v2')
const bestOf = ref<1 | 3 | 5>(1)

const team1Player1 = ref('')
const team1Player2 = ref('')
const team2Player1 = ref('')
const team2Player2 = ref('')

const canStartGame = computed(() => {
  if (gameFormat.value === '1v1') {
    return team1Player1.value.trim() && team2Player1.value.trim()
  } else {
    return (
      team1Player1.value.trim() &&
      team1Player2.value.trim() &&
      team2Player1.value.trim() &&
      team2Player2.value.trim()
    )
  }
})

const playerDbIds = ref<{ team1: (number | undefined)[], team2: (number | undefined)[] }>({
  team1: [],
  team2: []
})

function loadLeagueGamePlayers() {
  if (!leagueId || !leagueGameId) return

  const format = route.query.format as string
  const best = parseInt(route.query.bestOf as string)
  const t1p1 = route.query.t1p1 as string || ''
  const t1p2 = route.query.t1p2 as string || ''
  const t2p1 = route.query.t2p1 as string || ''
  const t2p2 = route.query.t2p2 as string || ''

  gameFormat.value = format === '1v1' ? '1v1' : '2v2'
  bestOf.value = (best === 3 || best === 5) ? best : 1
  team1Player1.value = t1p1
  team1Player2.value = t1p2
  team2Player1.value = t2p1
  team2Player2.value = t2p2

  const parseId = (val: string | undefined) => val ? parseInt(val) || undefined : undefined
  playerDbIds.value = {
    team1: [parseId(route.query.t1p1id as string), parseId(route.query.t1p2id as string)],
    team2: [parseId(route.query.t2p1id as string), parseId(route.query.t2p2id as string)]
  }
}

async function startGame() {
  const playerNames = {
    team1:
      gameFormat.value === '1v1'
        ? [team1Player1.value.trim()]
        : [team1Player1.value.trim(), team1Player2.value.trim()],
    team2:
      gameFormat.value === '1v1'
        ? [team2Player1.value.trim()]
        : [team2Player1.value.trim(), team2Player2.value.trim()]
  }

  playerStore.setupGame(gameFormat.value, playerNames, isLeagueGame ? playerDbIds.value : undefined)

  // Pickup games are local-only — no backend storage
  if (!isLeagueGame) {
    router.push('/games/pickup')
    return
  }

  try {
    const game: any = await api.fetch('/games', {
      method: 'POST',
      body: {
        format: gameFormat.value,
        bestOf: bestOf.value,
        players: playerNames,
        leagueGameId: parseInt(leagueGameId!)
      }
    })

    router.push(`/games/${game.id}?leagueId=${leagueId}`)
  } catch (err: any) {
    toast.error(err.data?.error || err.message)
  }
}

onMounted(() => {
  loadLeagueGamePlayers()
})
</script>

<template>
  <div class="max-w-2xl mx-auto p-6 space-y-6">

    <div class="text-center">
      <h1 class="text-3xl font-bold mb-2">Setup Game</h1>
      <p class="text-gray-600">{{ isLeagueGame ? 'League game — players pre-filled from schedule' : 'Choose format and enter player names' }}</p>
    </div>

    <!-- Game Format -->
    <div class="bg-white rounded-xl border shadow p-6">
      <h2 class="text-lg font-semibold mb-4">Game Format</h2>

      <div class="flex gap-4">
        <button @click="!isLeagueGame && (gameFormat = '1v1')" class="flex-1 py-3 px-6 rounded-lg font-semibold transition-all" :class="[
          gameFormat === '1v1' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700',
          isLeagueGame ? 'opacity-60 cursor-not-allowed' : 'hover:bg-gray-200'
        ]">
          1v1
        </button>

        <button @click="!isLeagueGame && (gameFormat = '2v2')" class="flex-1 py-3 px-6 rounded-lg font-semibold transition-all" :class="[
          gameFormat === '2v2' ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700',
          isLeagueGame ? 'opacity-60 cursor-not-allowed' : 'hover:bg-gray-200'
        ]">
          2v2
        </button>
      </div>
    </div>

    <!-- Best Of -->
    <div class="bg-white rounded-xl border shadow p-6">
      <h2 class="text-lg font-semibold mb-4">Match Format</h2>

      <div class="flex gap-4">
        <button v-for="n in [1, 3, 5]" :key="n" @click="!isLeagueGame && (bestOf = n as 1 | 3 | 5)"
          class="flex-1 py-3 px-6 rounded-lg font-semibold transition-all" :class="[
            bestOf === n ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-700',
            isLeagueGame ? 'opacity-60 cursor-not-allowed' : 'hover:bg-gray-200'
          ]">
          {{ n === 1 ? 'Single Game' : `Best of ${n}` }}
        </button>
      </div>
    </div>

    <!-- Players -->
    <div class="grid grid-cols-2 gap-6">

      <!-- Team 1 -->
      <div class="bg-red-50 rounded-xl border-2 border-red-200 p-6">
        <h2 class="text-lg font-semibold mb-4 text-red-700">Team 1</h2>

        <div class="space-y-3">
          <input v-model="team1Player1" type="text" placeholder="Player 1 Name" :readonly="isLeagueGame"
            class="w-full px-4 py-2 rounded-lg border border-red-300 focus:outline-none focus:ring-2 focus:ring-red-500"
            :class="{ 'bg-red-100 opacity-75 cursor-not-allowed': isLeagueGame }" />

          <input v-if="gameFormat === '2v2'" v-model="team1Player2" type="text" placeholder="Player 2 Name" :readonly="isLeagueGame"
            class="w-full px-4 py-2 rounded-lg border border-red-300 focus:outline-none focus:ring-2 focus:ring-red-500"
            :class="{ 'bg-red-100 opacity-75 cursor-not-allowed': isLeagueGame }" />
        </div>
      </div>

      <!-- Team 2 -->
      <div class="bg-blue-50 rounded-xl border-2 border-blue-200 p-6">
        <h2 class="text-lg font-semibold mb-4 text-blue-700">Team 2</h2>

        <div class="space-y-3">
          <input v-model="team2Player1" type="text" placeholder="Player 1 Name" :readonly="isLeagueGame"
            class="w-full px-4 py-2 rounded-lg border border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            :class="{ 'bg-blue-100 opacity-75 cursor-not-allowed': isLeagueGame }" />

          <input v-if="gameFormat === '2v2'" v-model="team2Player2" type="text" placeholder="Player 2 Name" :readonly="isLeagueGame"
            class="w-full px-4 py-2 rounded-lg border border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
            :class="{ 'bg-blue-100 opacity-75 cursor-not-allowed': isLeagueGame }" />
        </div>
      </div>

    </div>

    <!-- Start Game -->
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