<script setup lang="ts">
import { useToast } from '~/composables/useToast'

const toast = useToast()
</script>

<template>
    <Teleport to="body">
        <div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
            <TransitionGroup
                enter-active-class="transition-all duration-300 ease-out"
                enter-from-class="opacity-0 translate-y-2"
                enter-to-class="opacity-100 translate-y-0"
                leave-active-class="transition-all duration-200 ease-in"
                leave-from-class="opacity-100 translate-y-0"
                leave-to-class="opacity-0 translate-y-2"
            >
                <div
                    v-for="t in toast.toasts.value"
                    :key="t.id"
                    class="flex items-start gap-3 bg-white rounded-lg shadow-lg border-l-4 px-4 py-3 pointer-events-auto"
                    :class="{
                        'border-red-500': t.type === 'error',
                        'border-green-500': t.type === 'success',
                        'border-blue-500': t.type === 'info',
                    }"
                >
                    <span class="text-lg leading-none mt-0.5">
                        <span v-if="t.type === 'error'">✕</span>
                        <span v-else-if="t.type === 'success'">✓</span>
                        <span v-else>ℹ</span>
                    </span>
                    <p class="flex-1 text-sm text-gray-800 leading-snug">{{ t.message }}</p>
                    <button
                        @click="toast.dismiss(t.id)"
                        class="text-gray-400 hover:text-gray-600 leading-none text-lg font-bold flex-shrink-0"
                    >&times;</button>
                </div>
            </TransitionGroup>
        </div>
    </Teleport>
</template>
