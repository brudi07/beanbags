<script setup lang="ts">
import { ref } from 'vue';

const games = ref([
  {
    id: 1,
    team1: "Mike",
    team2: "Jake",
    score: "21-18",
    date: "Mar 5"
  },
  {
    id: 2,
    team1: "Mike",
    team2: "Sam",
    score: "21-10",
    date: "Mar 4"
  }
])
</script>

<template>
  <div class="max-w-5xl mx-auto p-6">

    <h1 class="text-3xl font-bold mb-8">
      Game History
    </h1>

    <div class="space-y-4">

      <NuxtLink
        v-for="game in games"
        :key="game.id"
        :to="`/game/${game.id}`"
        class="block border rounded-xl p-4 hover:bg-gray-50"
      >
        <div class="flex justify-between items-center">

          <div>
            <p class="font-semibold">
              {{ game.team1 }} vs {{ game.team2 }}
            </p>
            <p class="text-sm text-gray-500">
              {{ game.date }}
            </p>
          </div>

          <div class="text-xl font-bold">
            {{ game.score }}
          </div>

        </div>
      </NuxtLink>

    </div>

  </div>
</template>