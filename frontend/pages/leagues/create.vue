<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '~/composables/useAuth'
import type { CreateLeagueData } from '~/types/league'

const router = useRouter()
const auth = useAuth()

// Redirect if not organizer
if (!auth.canCreateLeague.value) {
    router.push('/')
}

const formData = ref<CreateLeagueData>({
    name: '',
    description: '',
    format: '2v2',
    maxTeams: 8,
    startDate: '',
    endDate: '',
    location: '',
    isPublic: true
})

const isLoading = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

async function handleCreateLeague() {
    if (!formData.value.name || !formData.value.location || !formData.value.startDate) {
        error.value = 'Please fill in all required fields'
        return
    }

    isLoading.value = true
    error.value = null
    success.value = false

    try {
        const response = await fetch('/api/leagues', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            },
            body: JSON.stringify(formData.value)
        })

        if (!response.ok) {
            const data = await response.json()
            throw new Error(data.message || 'Failed to create league')
        }

        const league = await response.json()
        success.value = true

        // Redirect to league page after short delay
        setTimeout(() => {
            router.push(`/leagues/${league.id}`)
        }, 1500)
    } catch (err: any) {
        error.value = err.message || 'Failed to create league. Please try again.'
    } finally {
        isLoading.value = false
    }
}
</script>

<template>
    <div class="max-w-3xl mx-auto py-8 px-4">

        <!-- Header -->
        <div class="mb-8">
            <h1 class="text-4xl font-bold text-gray-900 mb-2">Create New League</h1>
            <p class="text-gray-600">Set up a new cornhole league and start inviting players</p>
        </div>

        <!-- Success Message -->
        <div v-if="success" class="mb-6 bg-green-50 border border-green-200 rounded-lg p-4">
            <p class="text-green-700 font-semibold">✓ League created successfully! Redirecting...</p>
        </div>

        <!-- Form -->
        <div class="bg-white rounded-xl shadow-lg p-8">

            <form @submit.prevent="handleCreateLeague" class="space-y-6">

                <!-- League Name -->
                <div>
                    <label for="name" class="block text-sm font-semibold text-gray-700 mb-2">
                        League Name <span class="text-red-500">*</span>
                    </label>
                    <input id="name" v-model="formData.name" type="text" required
                        class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        placeholder="e.g., Summer 2024 Cornhole League" />
                </div>

                <!-- Description -->
                <div>
                    <label for="description" class="block text-sm font-semibold text-gray-700 mb-2">
                        Description
                    </label>
                    <textarea id="description" v-model="formData.description" rows="4"
                        class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        placeholder="Tell players about your league, rules, schedule, etc." />
                </div>

                <!-- Format and Max Teams -->
                <div class="grid grid-cols-2 gap-4">

                    <!-- Format -->
                    <div>
                        <label for="format" class="block text-sm font-semibold text-gray-700 mb-2">
                            Game Format <span class="text-red-500">*</span>
                        </label>
                        <select id="format" v-model="formData.format"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                            <option value="1v1">1v1 (Singles)</option>
                            <option value="2v2">2v2 (Doubles)</option>
                        </select>
                    </div>

                    <!-- Max Teams -->
                    <div>
                        <label for="maxTeams" class="block text-sm font-semibold text-gray-700 mb-2">
                            Max Teams <span class="text-red-500">*</span>
                        </label>
                        <input id="maxTeams" v-model.number="formData.maxTeams" type="number" min="2" max="64" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                    </div>

                </div>

                <!-- Start and End Dates -->
                <div class="grid grid-cols-2 gap-4">

                    <!-- Start Date -->
                    <div>
                        <label for="startDate" class="block text-sm font-semibold text-gray-700 mb-2">
                            Start Date <span class="text-red-500">*</span>
                        </label>
                        <input id="startDate" v-model="formData.startDate" type="date" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                    </div>

                    <!-- End Date -->
                    <div>
                        <label for="endDate" class="block text-sm font-semibold text-gray-700 mb-2">
                            End Date (Optional)
                        </label>
                        <input id="endDate" v-model="formData.endDate" type="date"
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                    </div>

                </div>

                <!-- Location -->
                <div>
                    <label for="location" class="block text-sm font-semibold text-gray-700 mb-2">
                        Location <span class="text-red-500">*</span>
                    </label>
                    <input id="location" v-model="formData.location" type="text" required
                        class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        placeholder="e.g., Joe's Backyard, 123 Main St, Eau Claire, WI" />
                </div>

                <!-- Public/Private -->
                <div class="bg-gray-50 rounded-lg p-4">
                    <label class="flex items-center cursor-pointer">
                        <input v-model="formData.isPublic" type="checkbox"
                            class="w-5 h-5 text-blue-600 rounded focus:ring-blue-500" />
                        <div class="ml-3">
                            <span class="font-semibold text-gray-900">Public League</span>
                            <p class="text-sm text-gray-600">Allow anyone to search and join this league</p>
                        </div>
                    </label>
                </div>

                <!-- Error Message -->
                <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
                    <p class="text-red-600 text-sm">{{ error }}</p>
                </div>

                <!-- Buttons -->
                <div class="flex gap-4">
                    <button type="submit" :disabled="isLoading || success"
                        class="flex-1 py-3 px-6 rounded-lg font-semibold text-white transition-colors" :class="isLoading || success
                            ? 'bg-gray-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'">
                        <span v-if="isLoading">Creating...</span>
                        <span v-else-if="success">Created!</span>
                        <span v-else>Create League</span>
                    </button>

                    <button type="button" @click="router.push('/leagues')"
                        class="px-6 py-3 rounded-lg font-semibold text-gray-700 border-2 border-gray-300 hover:bg-gray-50 transition-colors">
                        Cancel
                    </button>
                </div>

            </form>

        </div>

    </div>
</template>