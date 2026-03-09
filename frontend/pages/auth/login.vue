<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const username = ref('')
const password = ref('')
const isLoading = ref(false)
const error = ref<string | null>(null)

async function handleLogin() {
    if (!username.value || !password.value) {
        error.value = 'Please enter both username and password'
        return
    }

    isLoading.value = true
    error.value = null

    try {
        // TODO Replace with your actual API endpoint
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username.value,
                password: password.value,
            })
        })

        if (!response.ok) {
            const data = await response.json()
            throw new Error(data.message || 'Invalid credentials')
        }

        const data = await response.json()

        // Store auth token
        localStorage.setItem('authToken', data.token)

        // Redirect to home or dashboard
        router.push('/')
    } catch (err: any) {
        error.value = err.message || 'Login failed. Please try again.'
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
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Cornhole League</h1>
                <p class="text-gray-600">Sign in to your account</p>
            </div>

            <!-- Login Card -->
            <div class="bg-white rounded-xl shadow-lg p-8">

                <form @submit.prevent="handleLogin" class="space-y-6">

                    <!-- Username Field -->
                    <div>
                        <label for="username" class="block text-sm font-medium text-gray-700 mb-2">
                            Username
                        </label>
                        <input id="username" v-model="username" type="text" autocomplete="username" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Enter your username" />
                    </div>

                    <!-- Password Field -->
                    <div>
                        <label for="password" class="block text-sm font-medium text-gray-700 mb-2">
                            Password
                        </label>
                        <input id="password" v-model="password" type="password" autocomplete="current-password" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Enter your password" />
                    </div>

                    <!-- Error Message -->
                    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p class="text-red-600 text-sm">{{ error }}</p>
                    </div>

                    <!-- Forgot Password Link -->
                    <div class="text-right">
                        <NuxtLink to="/auth/forgot-password" class="text-sm text-blue-600 hover:text-blue-700 font-medium">
                            Forgot password?
                        </NuxtLink>
                    </div>

                    <!-- Login Button -->
                    <button type="submit" :disabled="isLoading"
                        class="w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors" :class="isLoading
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'">
                        <span v-if="isLoading">Signing in...</span>
                        <span v-else>Sign In</span>
                    </button>

                </form>

                <!-- Divider -->
                <div class="relative my-6">
                    <div class="absolute inset-0 flex items-center">
                        <div class="w-full border-t border-gray-300"></div>
                    </div>
                    <div class="relative flex justify-center text-sm">
                        <span class="px-2 bg-white text-gray-500">New to Cornhole League?</span>
                    </div>
                </div>

                <!-- Sign Up Link -->
                <div class="text-center">
                    <NuxtLink to="/auth/signup"
                        class="inline-block w-full py-3 px-4 rounded-lg font-semibold text-blue-600 border-2 border-blue-600 hover:bg-blue-50 transition-colors">
                        Create an account
                    </NuxtLink>
                </div>

            </div>

            <!-- Footer Links -->
            <div class="mt-6 text-center text-sm text-gray-600">
                <p>
                    By signing in, you agree to our
                    <NuxtLink to="/terms" class="text-blue-600 hover:text-blue-700">Terms of Service</NuxtLink>
                    and
                    <NuxtLink to="/privacy" class="text-blue-600 hover:text-blue-700">Privacy Policy</NuxtLink>
                </p>
            </div>

        </div>
    </div>
</template>