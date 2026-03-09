<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const formData = ref({
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
})

const isLoading = ref(false)
const error = ref<string | null>(null)

async function handleSignup() {
    // Validation
    if (!formData.value.username || !formData.value.email || !formData.value.password || !formData.value.confirmPassword) {
        error.value = 'Please fill in all fields'
        return
    }

    if (formData.value.password !== formData.value.confirmPassword) {
        error.value = 'Passwords do not match'
        return
    }

    if (formData.value.password.length < 6) {
        error.value = 'Password must be at least 6 characters'
        return
    }

    isLoading.value = true
    error.value = null

    try {
        // TODO Replace with your actual API endpoint
        const response = await fetch('/api/auth/signup', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: formData.value.username,
                email: formData.value.email,
                password: formData.value.password,
            })
        })

        if (!response.ok) {
            const data = await response.json()
            throw new Error(data.message || 'Signup failed')
        }

        const data = await response.json()

        // Store auth token
        localStorage.setItem('authToken', data.token)
        localStorage.setItem('username', formData.value.username)

        // Redirect to home
        router.push('/')
    } catch (err: any) {
        error.value = err.message || 'Signup failed. Please try again.'
    } finally {
        isLoading.value = false
    }
}
</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4 py-12">
        <div class="max-w-md w-full">

            <!-- Logo/Title -->
            <div class="text-center mb-8">
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Join Cornhole League</h1>
                <p class="text-gray-600">Create your account to start playing</p>
            </div>

            <!-- Signup Card -->
            <div class="bg-white rounded-xl shadow-lg p-8">

                <form @submit.prevent="handleSignup" class="space-y-5">

                    <!-- Username Field -->
                    <div>
                        <label for="username" class="block text-sm font-medium text-gray-700 mb-2">
                            Username
                        </label>
                        <input id="username" v-model="formData.username" type="text" autocomplete="username" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Choose a username" />
                    </div>

                    <!-- Email Field -->
                    <div>
                        <label for="email" class="block text-sm font-medium text-gray-700 mb-2">
                            Email Address
                        </label>
                        <input id="email" v-model="formData.email" type="email" autocomplete="email" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Enter your email" />
                    </div>

                    <!-- Password Field -->
                    <div>
                        <label for="password" class="block text-sm font-medium text-gray-700 mb-2">
                            Password
                        </label>
                        <input id="password" v-model="formData.password" type="password" autocomplete="new-password"
                            required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Create a password (min. 6 characters)" />
                    </div>

                    <!-- Confirm Password Field -->
                    <div>
                        <label for="confirmPassword" class="block text-sm font-medium text-gray-700 mb-2">
                            Confirm Password
                        </label>
                        <input id="confirmPassword" v-model="formData.confirmPassword" type="password"
                            autocomplete="new-password" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            placeholder="Confirm your password" />
                    </div>

                    <!-- Error Message -->
                    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-3">
                        <p class="text-red-600 text-sm">{{ error }}</p>
                    </div>

                    <!-- Signup Button -->
                    <button type="submit" :disabled="isLoading"
                        class="w-full py-3 px-4 rounded-lg font-semibold text-white transition-colors" :class="isLoading
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'">
                        <span v-if="isLoading">Creating account...</span>
                        <span v-else>Create Account</span>
                    </button>

                </form>

                <!-- Divider -->
                <div class="relative my-6">
                    <div class="absolute inset-0 flex items-center">
                        <div class="w-full border-t border-gray-300"></div>
                    </div>
                    <div class="relative flex justify-center text-sm">
                        <span class="px-2 bg-white text-gray-500">Already have an account?</span>
                    </div>
                </div>

                <!-- Sign In Link -->
                <div class="text-center">
                    <NuxtLink to="/auth/login"
                        class="inline-block w-full py-3 px-4 rounded-lg font-semibold text-blue-600 border-2 border-blue-600 hover:bg-blue-50 transition-colors">
                        Sign In
                    </NuxtLink>
                </div>

            </div>

        </div>
    </div>
</template>