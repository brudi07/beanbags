<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '~/composables/useAuth'
import type { League } from '~/types/league'

const router = useRouter()
const auth = useAuth()

const publicLeagues = ref<League[]>([])
const myLeagues = ref<League[]>([])
const isLoadingPublic = ref(true)
const isLoadingMy = ref(true)
const error = ref<string | null>(null)

// Browse filters
const searchQuery = ref('')
const formatFilter = ref<'all' | '1v1' | '2v2'>('all')
const statusFilter = ref<'all' | 'upcoming' | 'active'>('active') // Default to active

// Filter public leagues (exclude ones user is already in)
const availableLeagues = computed(() => {
    const myLeagueIds = new Set(myLeagues.value.map(l => l.id))

    return publicLeagues.value
        .filter(league => !myLeagueIds.has(league.id)) // Exclude leagues user is in
        .filter(league => {
            // Search filter
            if (searchQuery.value) {
                const query = searchQuery.value.toLowerCase()
                const matchesSearch =
                    league.name.toLowerCase().includes(query) ||
                    league.description.toLowerCase().includes(query) ||
                    league.location.toLowerCase().includes(query)
                if (!matchesSearch) return false
            }

            // Format filter
            if (formatFilter.value !== 'all' && league.format !== formatFilter.value) {
                return false
            }

            // Status filter
            if (statusFilter.value !== 'all' && league.status !== statusFilter.value) {
                return false
            }

            return true
        })
})

// Separate my leagues
const organizingLeagues = computed(() =>
    myLeagues.value.filter(league => league.organizerId === auth.currentUser.value?.id)
)

const playingLeagues = computed(() =>
    myLeagues.value.filter(league => league.organizerId !== auth.currentUser.value?.id)
)

// Quick stats
const activeLeagues = computed(() =>
    myLeagues.value.filter(league => league.status === 'active')
)

const upcomingLeagues = computed(() =>
    myLeagues.value.filter(league => league.status === 'upcoming')
)

const completedLeagues = computed(() =>
    myLeagues.value.filter(league => league.status === 'completed')
)

async function fetchPublicLeagues() {
    isLoadingPublic.value = true

    try {
        const response = await fetch('/api/leagues/public', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to fetch public leagues')
        publicLeagues.value = await response.json()
    } catch (err: any) {
        error.value = err.message
    } finally {
        isLoadingPublic.value = false
    }
}

async function fetchMyLeagues() {
    isLoadingMy.value = true

    try {
        const response = await fetch('/api/leagues/my-leagues', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to fetch your leagues')
        myLeagues.value = await response.json()
    } catch (err: any) {
        error.value = err.message
    } finally {
        isLoadingMy.value = false
    }
}

async function joinLeague(leagueId: string) {
    try {
        const response = await fetch(`/api/leagues/${leagueId}/join`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to join league')

        // Refresh both lists
        await Promise.all([fetchPublicLeagues(), fetchMyLeagues()])

        // Redirect to league page
        router.push(`/leagues/${leagueId}`)
    } catch (err: any) {
        alert(err.message || 'Failed to join league')
    }
}

async function leaveLeague(leagueId: string) {
    if (!confirm('Are you sure you want to leave this league?')) return

    try {
        const response = await fetch(`/api/leagues/${leagueId}/leave`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to leave league')

        // Refresh both lists
        await Promise.all([fetchPublicLeagues(), fetchMyLeagues()])
    } catch (err: any) {
        alert(err.message || 'Failed to leave league')
    }
}

function viewLeague(leagueId: string) {
    router.push(`/leagues/${leagueId}`)
}

function createLeague() {
    router.push('/leagues/create')
}

onMounted(() => {
    fetchPublicLeagues()
    fetchMyLeagues()
})
</script>

<template>
    <div class="max-w-6xl mx-auto py-8 px-4">

        <!-- ========== BROWSE & JOIN SECTION ========== -->
        <div class="mb-12">

            <!-- Header -->
            <div class="mb-6">
                <h1 class="text-3xl font-bold text-gray-900 mb-2">Browse & Join Leagues</h1>
                <p class="text-gray-600">Find and join cornhole leagues in your area</p>
            </div>

            <!-- Filters -->
            <div class="bg-white rounded-xl shadow-sm p-6 mb-6">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">

                    <!-- Search -->
                    <div class="md:col-span-1">
                        <label for="search" class="block text-sm font-medium text-gray-700 mb-2">
                            Search
                        </label>
                        <input id="search" v-model="searchQuery" type="text" placeholder="League name, location..."
                            class="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />
                    </div>

                    <!-- Format Filter -->
                    <div>
                        <label for="format" class="block text-sm font-medium text-gray-700 mb-2">
                            Format
                        </label>
                        <select id="format" v-model="formatFilter"
                            class="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
                            <option value="all">All Formats</option>
                            <option value="1v1">1v1 Only</option>
                            <option value="2v2">2v2 Only</option>
                        </select>
                    </div>

                    <!-- Status Filter -->
                    <div>
                        <label for="status" class="block text-sm font-medium text-gray-700 mb-2">
                            Status
                        </label>
                        <select id="status" v-model="statusFilter"
                            class="w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500">
                            <option value="all">All</option>
                            <option value="upcoming">Upcoming</option>
                            <option value="active">Active</option>
                        </select>
                    </div>

                </div>
            </div>

            <!-- Loading State -->
            <div v-if="isLoadingPublic" class="text-center py-12">
                <p class="text-gray-500">Loading leagues...</p>
            </div>

            <!-- Empty State -->
            <div v-else-if="availableLeagues.length === 0" class="text-center py-12 bg-gray-50 rounded-xl">
                <p class="text-gray-500 text-lg mb-2">No available leagues found</p>
                <p class="text-gray-400 text-sm">
                    {{ searchQuery || formatFilter !== 'all' || statusFilter !== 'all'
                        ? 'Try adjusting your filters'
                        : 'Check back later for new leagues' }}
                </p>
            </div>

            <!-- Leagues Grid -->
            <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <div v-for="league in availableLeagues" :key="league.id"
                    class="bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow overflow-hidden">

                    <!-- Header -->
                    <div class="px-6 pt-6 pb-4 border-b border-gray-100">
                        <div class="flex items-start justify-between mb-3">
                            <h3 class="text-xl font-bold text-gray-900 flex-1">{{ league.name }}</h3>
                            <span class="px-3 py-1 rounded-full text-xs font-semibold ml-2" :class="{
                                'bg-green-100 text-green-700': league.status === 'active',
                                'bg-blue-100 text-blue-700': league.status === 'upcoming',
                            }">
                                {{ league.status }}
                            </span>
                        </div>
                        <p class="text-sm text-gray-600 line-clamp-2">{{ league.description }}</p>
                    </div>

                    <!-- Info -->
                    <div class="px-6 py-4 space-y-2 text-sm">
                        <div class="flex items-center text-gray-600">
                            <span class="font-medium w-24">Format:</span>
                            <span>{{ league.format }}</span>
                        </div>
                        <div class="flex items-center text-gray-600">
                            <span class="font-medium w-24">Teams:</span>
                            <span>{{ league.currentTeams }}/{{ league.maxTeams }}</span>
                        </div>
                        <div class="flex items-center text-gray-600">
                            <span class="font-medium w-24">Location:</span>
                            <span class="truncate">{{ league.location }}</span>
                        </div>
                        <div class="flex items-center text-gray-600">
                            <span class="font-medium w-24">Starts:</span>
                            <span>{{ new Date(league.startDate).toLocaleDateString() }}</span>
                        </div>
                    </div>

                    <!-- Actions -->
                    <div class="px-6 pb-6 flex gap-2">
                        <button @click="viewLeague(league.id)"
                            class="flex-1 py-2 px-4 rounded-lg font-semibold text-gray-700 border-2 border-gray-300 hover:bg-gray-50 transition-colors">
                            View
                        </button>
                        <button v-if="league.currentTeams < league.maxTeams" @click="joinLeague(league.id)"
                            class="flex-1 py-2 px-4 rounded-lg font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors">
                            Join
                        </button>
                    </div>

                </div>
            </div>

        </div>

        <!-- ========== MY LEAGUES SECTION ========== -->
        <div class="border-t-4 border-gray-200 pt-12">

            <!-- Header -->
            <div class="flex items-center justify-between mb-6">
                <div>
                    <h2 class="text-3xl font-bold text-gray-900 mb-2">My Leagues</h2>
                    <p class="text-gray-600">Leagues you're organizing or playing in</p>
                </div>

                <button v-if="auth.canCreateLeague.value" @click="createLeague"
                    class="px-6 py-3 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 transition-colors">
                    + Create League
                </button>
            </div>

            <!-- Loading State -->
            <div v-if="isLoadingMy" class="text-center py-12">
                <p class="text-gray-500">Loading your leagues...</p>
            </div>

            <!-- Empty State -->
            <div v-else-if="myLeagues.length === 0" class="text-center py-12 bg-gray-50 rounded-xl">
                <p class="text-gray-500 text-lg mb-4">You're not part of any leagues yet</p>
                <p class="text-gray-400 text-sm mb-6">Join a league above or create your own!</p>
            </div>

            <!-- My Leagues Content -->
            <div v-else class="space-y-8">

                <!-- Quick Stats -->
                <div class="bg-gradient-to-br from-blue-50 to-blue-100 rounded-xl p-6">
                    <h3 class="text-lg font-bold text-gray-900 mb-4">Quick Stats</h3>
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <div class="bg-white rounded-lg p-4 text-center">
                            <p class="text-3xl font-bold text-blue-600">{{ activeLeagues.length }}</p>
                            <p class="text-sm text-gray-600">Active</p>
                        </div>
                        <div class="bg-white rounded-lg p-4 text-center">
                            <p class="text-3xl font-bold text-green-600">{{ upcomingLeagues.length }}</p>
                            <p class="text-sm text-gray-600">Upcoming</p>
                        </div>
                        <div class="bg-white rounded-lg p-4 text-center">
                            <p class="text-3xl font-bold text-gray-600">{{ completedLeagues.length }}</p>
                            <p class="text-sm text-gray-600">Completed</p>
                        </div>
                        <div class="bg-white rounded-lg p-4 text-center">
                            <p class="text-3xl font-bold text-purple-600">{{ myLeagues.length }}</p>
                            <p class="text-sm text-gray-600">Total</p>
                        </div>
                    </div>
                </div>

                <!-- Organizing Leagues -->
                <div v-if="organizingLeagues.length > 0">
                    <h3 class="text-2xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <span>Leagues I Organize</span>
                        <span class="text-sm font-normal text-gray-500">({{ organizingLeagues.length }})</span>
                    </h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        <LeagueCard v-for="league in organizingLeagues" :key="league.id" :league="league"
                            :is-organizer="true" @view="viewLeague" />
                    </div>
                </div>

                <!-- Playing Leagues -->
                <div v-if="playingLeagues.length > 0">
                    <h3 class="text-2xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <span>Leagues I'm Playing In</span>
                        <span class="text-sm font-normal text-gray-500">({{ playingLeagues.length }})</span>
                    </h3>
                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        <LeagueCard v-for="league in playingLeagues" :key="league.id" :league="league"
                            :is-organizer="false" @view="viewLeague" @leave="leaveLeague" />
                    </div>
                </div>

            </div>

        </div>

    </div>
</template>

<!-- League Card Component -->
<script lang="ts">
export default {
    components: {
        LeagueCard: {
            props: {
                league: Object,
                isOrganizer: Boolean
            },
            emits: ['view', 'leave'],
            template: `
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow overflow-hidden">

          <div class="px-6 pt-6 pb-4 border-b border-gray-100">
            <div class="flex items-start justify-between mb-3">
              <h4 class="text-xl font-bold text-gray-900 flex-1">{{ league.name }}</h4>
              <span
                class="px-3 py-1 rounded-full text-xs font-semibold ml-2"
                :class="{
                  'bg-green-100 text-green-700': league.status === 'active',
                  'bg-blue-100 text-blue-700': league.status === 'upcoming',
                  'bg-gray-100 text-gray-700': league.status === 'completed'
                }"
              >
                {{ league.status }}
              </span>
            </div>
            <p class="text-sm text-gray-600 line-clamp-2">{{ league.description }}</p>
          </div>

          <div class="px-6 py-4 space-y-2 text-sm">
            <div class="flex items-center text-gray-600">
              <span class="font-medium w-24">Format:</span>
              <span>{{ league.format }}</span>
            </div>
            <div class="flex items-center text-gray-600">
              <span class="font-medium w-24">Teams:</span>
              <span>{{ league.currentTeams }}/{{ league.maxTeams }}</span>
            </div>
            <div class="flex items-center text-gray-600">
              <span class="font-medium w-24">Location:</span>
              <span class="truncate">{{ league.location }}</span>
            </div>
            <div v-if="isOrganizer" class="flex items-center">
              <span class="px-2 py-1 bg-purple-100 text-purple-700 text-xs font-semibold rounded">
                Organizer
              </span>
            </div>
          </div>

          <div class="px-6 pb-6 flex gap-2">
            <button
              @click="$emit('view', league.id)"
              class="flex-1 py-2 px-4 rounded-lg font-semibold text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              {{ isOrganizer ? 'Manage' : 'View' }}
            </button>
            <button
              v-if="!isOrganizer"
              @click="$emit('leave', league.id)"
              class="py-2 px-4 rounded-lg font-semibold text-red-600 border-2 border-red-600 hover:bg-red-50 transition-colors"
            >
              Leave
            </button>
          </div>

        </div>
      `
        }
    }
}
</script>