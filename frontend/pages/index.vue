<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen flex flex-col items-center justify-center gap-8">

    <h1 class="text-5xl font-bold">
      Bean Bag Tracker
    </h1>

    <p class="text-lg text-gray-500">
      Track scores, throws, and analytics for every game.
    </p>

    <div class="flex gap-4">

      <NuxtLink
        to="/games/new"
        class="bg-blue-600 text-white px-6 py-3 rounded-lg"
      >
        Start Pickup Game
      </NuxtLink>

    </div>

  </div>
</template>