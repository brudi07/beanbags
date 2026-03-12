<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '~/composables/useApi'

const route = useRoute()
const router = useRouter()
const api = useApi()

const token = route.query.token as string | undefined

const password = ref('')
const confirmPassword = ref('')
const isLoading = ref(false)
const success = ref(false)
const error = ref<string | null>(null)

const passwordMismatch = computed(() =>
    confirmPassword.value.length > 0 && password.value !== confirmPassword.value
)

const canSubmit = computed(() =>
    password.value.length >= 6 && password.value === confirmPassword.value
)

async function handleReset() {
    if (!canSubmit.value) return

    isLoading.value = true
    error.value = null

    try {
        await api.fetch('/auth/reset-password', {
            method: 'POST',
            body: { token, password: password.value }
        })

        success.value = true
    } catch (err: any) {
        error.value = err.data?.error || err.message || 'Failed to reset password. Please try again.'
    } finally {
        isLoading.value = false
    }
}
</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
        <div class="max-w-md w-full">

            <!-- Logo/Title -->
            <div class="text-center mb-8">
                <h1 class="text-4xl font-bold text-gray-900 mb-2">New Password</h1>
                <p class="text-gray-600">Enter a new password for your account</p>
            </div>

            <!-- Card -->
            <div class="bg-white rounded-xl shadow-lg p-8">

                <!-- Invalid link (no token) -->
                <div v-if="!token" class="text-center py-4">
                    <p class="text-red-600 mb-4">This reset link is invalid or missing a token.</p>
                    <NuxtLink to="/auth/forgot-password"
                        class="text-blue-600 hover:text-blue-700 font-medium text-sm">
                        Request a new reset link →
                    </NuxtLink>
                </div>

                <!-- Success -->
                <div v-else-if="success" class="text-center py-4">
                    <div class="mb-4 bg-green-50 border border-green-200 rounded-lg p-4">
                        <p class="text-green-700 text-sm">✓ Password updated successfully!</p>
                    </div>
                    <NuxtLink to="/auth/login"
                        class="inline-block px-6 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                        Log in with new password
                    </NuxtLink>
                </div>

                <!-- Reset form -->
                <form v-else @submit.prevent="handleReset" class="space-y-5">

                    <div>
                        <label for="password" class="block text-sm font-medium text-gray-700 mb-2">
                            New Password
                        </label>
                        <input id="password" v-model="password" type="password" autocomplete="new-password" required
                            minlength="6"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="At least 6 characters" />
                    </div>

                    <div>
                        <label for="confirm" class="block text-sm font-medium text-gray-700 mb-2">
                            Confirm Password
                        </label>
                        <input id="confirm" v-model="confirmPassword" type="password" autocomplete="new-password"
                            required
                            class="w-full px-4 py-3 rounded-lg border focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            :class="passwordMismatch ? 'border-red-400' : 'border-gray-300'"
                            placeholder="Repeat your new password" />
                        <p v-if="passwordMismatch" class="mt-1 text-xs text-red-500">Passwords do not match</p>
                    </div>

                    <!-- Error -->
                    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p class="text-red-600 text-sm">{{ error }}</p>
                    </div>

                    <button type="submit" :disabled="!canSubmit || isLoading"
                        class="w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors"
                        :class="canSubmit && !isLoading ? 'bg-blue-600 hover:bg-blue-700' : 'bg-gray-400 cursor-not-allowed'">
                        <span v-if="isLoading">Updating...</span>
                        <span v-else>Set New Password</span>
                    </button>

                </form>

                <!-- Back to login -->
                <div v-if="!success" class="mt-6 text-center">
                    <NuxtLink to="/auth/login" class="text-sm text-blue-600 hover:text-blue-700 font-medium">
                        ← Back to login
                    </NuxtLink>
                </div>

            </div>
        </div>
    </div>
</template>
