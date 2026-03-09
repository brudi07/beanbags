<script setup lang="ts">
import { computed } from 'vue';
import { useScoringStore } from '~/stores/scoringStore';

const scoringStore = useScoringStore()

const totalThrows = computed(() => scoringStore.throws.length)
</script>

<template>
  <div class="max-w-6xl mx-auto p-6">

    <h1 class="text-3xl font-bold mb-8">
      Player Stats
    </h1>

    <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">

      <div class="border rounded-xl p-6">
        <p class="text-sm text-gray-500">Total Throws</p>
        <p class="text-3xl font-bold">
          {{ totalThrows }}
        </p>
      </div>

      <div class="border rounded-xl p-6">
        <p class="text-sm text-gray-500">Board Hits</p>
        <p class="text-3xl font-bold">0</p>
      </div>

      <div class="border rounded-xl p-6">
        <p class="text-sm text-gray-500">Hole Shots</p>
        <p class="text-3xl font-bold">0</p>
      </div>

      <div class="border rounded-xl p-6">
        <p class="text-sm text-gray-500">Accuracy</p>
        <p class="text-3xl font-bold">0%</p>
      </div>

    </div>

  </div>
</template>