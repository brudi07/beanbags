<script setup lang="ts">
import { useScoringStore } from '~/stores/scoringStore'

const scoringStore = useScoringStore()
</script>

<template>
  <div class="grid grid-cols-3 gap-6">

    <ScoreBoard />

    <CornholeBoard />

    <RoundTracker />

  </div>
</template>