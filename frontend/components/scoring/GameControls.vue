<script setup lang="ts">
import { useScoringStore } from '~/stores/scoringStore';

const scoringStore = useScoringStore()
</script>

<template>
  <div class="flex flex-col md:flex-row gap-4 justify-center">

    <button
      @click="scoringStore.scoreRound()"
      class="bg-green-600 text-white px-6 py-3 rounded-lg"
    >
      Finish Round
    </button>

  </div>
</template>