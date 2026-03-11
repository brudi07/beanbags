<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import type { League, CreateLeagueData } from '~/types/league'

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const api = useApi()

const leagueId = route.params.id as string

const league = ref<League | null>(null)
const formData = ref<CreateLeagueData>({
    name: '',
    description: '',
    format: '2v2',
    games_per_match: 3,
    max_teams: 8,
    start_date: '',
    weeks_of_play: 8,
    location: '',
    is_public: true
})

const isLoading = ref(true)
const isSaving = ref(false)
const error = ref<string | null>(null)
const success = ref(false)

async function fetchLeague() {
    isLoading.value = true
    error.value = null

    try {
        league.value = await api.fetch<League>(`/leagues/${leagueId}`)

        // Check if user is organizer
        if (league.value.organizer_id !== auth.currentUser.value?.id) {
            error.value = 'Only the organizer can edit this league'
            setTimeout(() => router.push(`/leagues/${leagueId}`), 2000)
            return
        }

        // Populate form with league data
        formData.value = {
            name: league.value.name,
            description: league.value.description,
            format: league.value.format,
            games_per_match: league.value.games_per_match,
            max_teams: league.value.max_teams,
            start_date: league.value.start_date,
            weeks_of_play: league.value.weeks_of_play,
            location: league.value.location,
            is_public: league.value.is_public
        }
    } catch (err: any) {
        error.value = err.data?.error || err.message || 'Failed to load league'
    } finally {
        isLoading.value = false
    }
}

async function handleUpdateLeague() {
    if (!formData.value.name || !formData.value.location || !formData.value.start_date) {
        error.value = 'Please fill in all required fields'
        return
    }

    isSaving.value = true
    error.value = null
    success.value = false

    try {
        await api.fetch(`/leagues/${leagueId}`, {
            method: 'PATCH',
            body: formData.value
        })

        success.value = true

        // Redirect back to league page after short delay
        setTimeout(() => {
            router.push(`/leagues/${leagueId}`)
        }, 1500)
    } catch (err: any) {
        error.value = err.data?.error || err.message || 'Failed to update league. Please try again.'
    } finally {
        isSaving.value = false
    }
}

function cancel() {
    router.push(`/leagues/${leagueId}`)
}

onMounted(() => {
    fetchLeague()
})
</script>

<template>
    <div class="max-w-3xl mx-auto py-8 px-4">

        <!-- Loading State -->
        <div v-if="isLoading" class="text-center py-12">
            <p class="text-gray-500">Loading league...</p>
        </div>

        <!-- Error State (Permission) -->
        <div v-else-if="error && !league" class="bg-red-50 border border-red-200 rounded-lg p-4">
            <p class="text-red-600">{{ error }}</p>
        </div>

        <!-- Edit Form -->
        <div v-else>

            <!-- Header -->
            <div class="mb-8">
                <h1 class="text-4xl font-bold text-gray-900 mb-2">Edit League</h1>
                <p class="text-gray-600">Update your league settings and information</p>
            </div>

            <!-- Success Message -->
            <div v-if="success" class="mb-6 bg-green-50 border border-green-200 rounded-lg p-4">
                <p class="text-green-700 font-semibold">✓ League updated successfully! Redirecting...</p>
            </div>

            <!-- Form -->
            <div class="bg-white rounded-xl shadow-lg p-8">

                <form @submit.prevent="handleUpdateLeague" class="space-y-6">

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

                    <!-- Format and Games Per Match -->
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

                        <!-- Games Per Match -->
                        <div>
                            <label for="gamesPerMatch" class="block text-sm font-semibold text-gray-700 mb-2">
                                Match Format <span class="text-red-500">*</span>
                            </label>
                            <select id="gamesPerMatch" v-model.number="formData.games_per_match"
                                class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                                <option :value="1">Best of 1</option>
                                <option :value="3">Best of 3</option>
                                <option :value="5">Best of 5</option>
                            </select>
                        </div>

                    </div>

                    <!-- Max Teams -->
                    <div>
                        <label for="maxTeams" class="block text-sm font-semibold text-gray-700 mb-2">
                            Maximum Teams <span class="text-red-500">*</span>
                        </label>
                        <input id="maxTeams" v-model.number="formData.max_teams" type="number" min="2" max="64" required
                            class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                        <p v-if="league" class="text-xs text-gray-500 mt-1">
                            Current teams: {{ league.current_teams }}. You cannot set max below current.
                        </p>
                    </div>

                    <!-- Start Date and Weeks of Play -->
                    <div class="grid grid-cols-2 gap-4">

                        <!-- Start Date -->
                        <div>
                            <label for="startDate" class="block text-sm font-semibold text-gray-700 mb-2">
                                Start Date <span class="text-red-500">*</span>
                            </label>
                            <input id="startDate" v-model="formData.start_date" type="date" required
                                class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
                        </div>

                        <!-- Weeks of Play -->
                        <div>
                            <label for="weeksOfPlay" class="block text-sm font-semibold text-gray-700 mb-2">
                                Weeks of Play <span class="text-red-500">*</span>
                            </label>
                            <input id="weeksOfPlay" v-model.number="formData.weeks_of_play" type="number" min="1"
                                max="52" required
                                class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                placeholder="e.g., 8" />
                            <p class="text-xs text-gray-500 mt-1">League will end {{ formData.weeks_of_play }} weeks
                                after start date</p>
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
                            <input v-model="formData.is_public" type="checkbox"
                                class="w-5 h-5 text-blue-600 rounded focus:ring-blue-500" />
                            <div class="ml-3">
                                <span class="font-semibold text-gray-900">Public League</span>
                                <p class="text-sm text-gray-600">Allow anyone to search and join this league</p>
                            </div>
                        </label>
                    </div>

                    <!-- Warning about changes -->
                    <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                        <p class="text-yellow-800 text-sm font-semibold mb-1">⚠️ Important Notes:</p>
                        <ul class="text-yellow-700 text-sm space-y-1 ml-4 list-disc">
                            <li>Changing format or games per match may affect scheduled games</li>
                            <li>You cannot reduce max teams below current member count</li>
                            <li>Changing the start date may require rescheduling games</li>
                        </ul>
                    </div>

                    <!-- Error Message -->
                    <div v-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4">
                        <p class="text-red-600 text-sm">{{ error }}</p>
                    </div>

                    <!-- Buttons -->
                    <div class="flex gap-4">
                        <button type="submit" :disabled="isSaving || success"
                            class="flex-1 py-3 px-6 rounded-lg font-semibold text-white transition-colors" :class="isSaving || success
                                ? 'bg-gray-400 cursor-not-allowed'
                                : 'bg-blue-600 hover:bg-blue-700'">
                            <span v-if="isSaving">Saving...</span>
                            <span v-else-if="success">Saved!</span>
                            <span v-else>Save Changes</span>
                        </button>

                        <button type="button" @click="cancel" :disabled="isSaving || success"
                            class="px-6 py-3 rounded-lg font-semibold text-gray-700 border-2 border-gray-300 hover:bg-gray-50 transition-colors disabled:opacity-50">
                            Cancel
                        </button>
                    </div>

                </form>

            </div>

        </div>

    </div>
</template>