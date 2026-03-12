<script setup lang="ts">
import { useConfirm } from '~/composables/useConfirm'

const { state, respond } = useConfirm()
</script>

<template>
    <Teleport to="body">
        <Transition name="fade">
            <div v-if="state" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
                @click.self="respond(false)">
                <div class="bg-white rounded-xl shadow-2xl max-w-md w-full p-6">
                    <h3 class="text-lg font-bold text-gray-900 mb-2">{{ state.title }}</h3>
                    <p class="text-gray-600 mb-6">{{ state.message }}</p>
                    <div class="flex gap-3 justify-end">
                        <button @click="respond(false)"
                            class="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg font-semibold hover:bg-gray-200 transition-colors">
                            {{ state.cancelLabel }}
                        </button>
                        <button @click="respond(true)"
                            class="px-4 py-2 rounded-lg font-semibold transition-colors"
                            :class="state.danger
                                ? 'bg-red-600 text-white hover:bg-red-700'
                                : 'bg-blue-600 text-white hover:bg-blue-700'">
                            {{ state.confirmLabel }}
                        </button>
                    </div>
                </div>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}
</style>
