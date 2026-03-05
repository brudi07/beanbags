<template>
  <div
    class="bag"
    draggable="true"
    @dragend="handleDrop"
  />
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Bag',
  methods: {
    handleDrop(event: DragEvent) {
      // placeholder implementation to satisfy TypeScript; add real logic here
      event.preventDefault();
    },
  },
});
</script>