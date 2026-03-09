<script setup lang="ts">
defineProps<{
  team: 1 | 2
}>()
</script>

<template>
  <div class="w-12 h-12 rounded shadow-md cursor-grab" :class="team === 1 ? 'bg-blue-500' : 'bg-red-500'" />
</template>