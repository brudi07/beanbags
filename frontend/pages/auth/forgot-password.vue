<script setup lang="ts">
import { ref } from 'vue'

const email = ref('')
const isLoading = ref(false)
const success = ref(false)
const error = ref<string | null>(null)

async function handleResetPassword() {
    if (!email.value) {
        error.value = 'Please enter your email address'
        return
    }

    isLoading.value = true
    error.value = null
    success.value = false

    try {
        // TODO Replace with your actual API endpoint
        const response = await fetch('/api/auth/forgot-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                email: email.value,
            })
        })

        if (!response.ok) {
            throw new Error('Failed to send reset email')
        }

        success.value = true
        email.value = ''
    } catch (err: any) {
        error.value = err.message || 'Failed to send reset email. Please try again.'
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
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Reset Password</h1>
                <p class="text-gray-600">Enter your email to receive a password reset link</p>
            </div>

            <!-- Reset Card -->
            <div class="bg-white rounded-xl shadow-lg p-8">

                <!-- Success Message -->
                <div v-if="success" class="mb-6 bg-green-50 border border-green-200 rounded-lg p-4">
                    <p class="text-green-700 text-sm">
                        ✓ Password reset link sent! Check your email for instructions.
                    </p>
                </div>

                <form @submit.prevent="handleResetPassword" class="space-y-6">

                    <!-- Email Field -->
                    <div>
                        <label for="email" class="block text-sm font-medium text-gray-700 mb-2">
                            Email Address
                        </label>
                        <input id="email" v-model="email" type="email" autocomplete="email" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Enter your email" />
                    </div>

                    <!-- Error Message -->
                    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p class="text-red-600 text-sm">{{ error }}</p>
                    </div>

                    <!-- Submit Button -->
                    <button type="submit" :disabled="isLoading"
                        class="w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors" :class="isLoading
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'">
                        <span v-if="isLoading">Sending...</span>
                        <span v-else>Send Reset Link</span>
                    </button>

                </form>

                <!-- Back to Login -->
                <div class="mt-6 text-center">
                    <NuxtLink to="/auth/login" class="text-sm text-blue-600 hover:text-blue-700 font-medium">
                        ← Back to login
                    </NuxtLink>
                </div>

            </div>

        </div>
    </div>
</template>