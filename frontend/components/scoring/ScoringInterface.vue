<script setup lang="ts">
const team1Bags = [1, 2, 3, 4]
const team2Bags = [1, 2, 3, 4]
</script>

<template>

  <div class="grid grid-cols-1 md:grid-cols-3 gap-6">

    <!-- team1 bags -->
    <div class="flex flex-col items-center gap-3">

      <Bag v-for="n in team1Bags" :key="n" :team="1" />

    </div>

    <!-- board -->
    <CornholeBoard />

    <!-- team2 bags -->
    <div class="flex flex-col items-center gap-3">

      <Bag v-for="n in team2Bags" :key="n" :team="2" />

    </div>

  </div>

</template>