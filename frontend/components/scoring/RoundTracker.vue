<script setup lang="ts">
import { useScoringStore } from '~/stores/scoringStore'

const scoringStore = useScoringStore()
</script>

<template>

    <div class="border rounded-xl p-4 bg-white shadow max-h-64 overflow-y-auto">

        <h2 class="font-semibold mb-3">
            Round History
        </h2>

        <div v-for="round in scoringStore.roundHistory" :key="round.round"
            class="flex justify-between text-sm border-b py-1">

            <span>Round {{ round.round }}</span>

            <span>
                {{ round.team1Points }} - {{ round.team2Points }}
            </span>

        </div>

    </div>

</template>