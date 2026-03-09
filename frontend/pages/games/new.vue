<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter()

const team1 = ref('')
const team2 = ref('')

function startGame() {
  // later this will create a game id
  router.push('/games/1')
}
</script>

<template>
  <div class="max-w-3xl mx-auto px-4 py-8">

    <h1 class="text-3xl font-bold mb-8 text-center">
      Start New Game
    </h1>

    <div class="grid gap-6 md:grid-cols-2">

      <!-- Team 1 -->
      <div class="border rounded-xl p-6">
        <h2 class="font-semibold text-lg mb-4">Team 1</h2>

        <input
          v-model="team1"
          placeholder="Enter player name"
          class="w-full border rounded-lg p-3"
        />
      </div>

      <!-- Team 2 -->
      <div class="border rounded-xl p-6">
        <h2 class="font-semibold text-lg mb-4">Team 2</h2>

        <input
          v-model="team2"
          placeholder="Enter player name"
          class="w-full border rounded-lg p-3"
        />
      </div>

    </div>

    <div class="flex justify-center mt-10">
      <button
        @click="startGame"
        class="bg-blue-600 text-white px-8 py-4 rounded-xl text-lg w-full md:w-auto"
      >
        Start Game
      </button>
    </div>

  </div>
</template>