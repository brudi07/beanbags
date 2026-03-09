<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuth } from '~/composables/useAuth'
import type { League } from '~/types/league'

const route = useRoute()
const router = useRouter()
const auth = useAuth()

const leagueId = route.params.id as string

const league = ref<League | null>(null)
const members = ref<any[]>([])
const games = ref<any[]>([])
const standings = ref<any[]>([])

const isLoading = ref(true)
const error = ref<string | null>(null)
const activeTab = ref<'overview' | 'standings' | 'games' | 'members' | 'settings'>('overview')

const isOrganizer = computed(() =>
    league.value && auth.currentUser.value &&
    league.value.organizerId === auth.currentUser.value.id
)

const isMember = computed(() =>
    members.value.some(m => m.playerId === auth.currentUser.value?.id)
)

async function fetchLeagueData() {
    isLoading.value = true
    error.value = null

    try {
        const [leagueRes, membersRes, gamesRes, standingsRes] = await Promise.all([
            fetch(`/api/leagues/${leagueId}`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('authToken')}` }
            }),
            fetch(`/api/leagues/${leagueId}/members`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('authToken')}` }
            }),
            fetch(`/api/leagues/${leagueId}/games`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('authToken')}` }
            }),
            fetch(`/api/leagues/${leagueId}/standings`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('authToken')}` }
            })
        ])

        if (!leagueRes.ok) throw new Error('Failed to fetch league')

        league.value = await leagueRes.json()
        members.value = membersRes.ok ? await membersRes.json() : []
        games.value = gamesRes.ok ? await gamesRes.json() : []
        standings.value = standingsRes.ok ? await standingsRes.json() : []
    } catch (err: any) {
        error.value = err.message
    } finally {
        isLoading.value = false
    }
}

async function joinLeague() {
    try {
        const response = await fetch(`/api/leagues/${leagueId}/join`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to join league')
        await fetchLeagueData()
    } catch (err: any) {
        alert(err.message)
    }
}

async function leaveLeague() {
    if (!confirm('Are you sure you want to leave this league?')) return

    try {
        const response = await fetch(`/api/leagues/${leagueId}/leave`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('authToken')}`
            }
        })

        if (!response.ok) throw new Error('Failed to leave league')
        router.push('/leagues')
    } catch (err: any) {
        alert(err.message)
    }
}

function createGame() {
    router.push(`/games/new?leagueId=${leagueId}`)
}

function editLeague() {
    router.push(`/leagues/${leagueId}/edit`)
}

onMounted(() => {
    fetchLeagueData()
})
</script>

<template>
    <div class="max-w-6xl mx-auto py-8 px-4">

        <!-- Loading State -->
        <div v-if="isLoading" class="text-center py-12">
            <p class="text-gray-500">Loading league...</p>
        </div>

        <!-- Error State -->
        <div v-else-if="error || !league" class="bg-red-50 border border-red-200 rounded-lg p-4">
            <p class="text-red-600">{{ error || 'League not found' }}</p>
        </div>

        <!-- League Content -->
        <div v-else class="space-y-6">

            <!-- Header -->
            <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <div class="flex items-start justify-between mb-4">
                    <div class="flex-1">
                        <div class="flex items-center gap-3 mb-2">
                            <h1 class="text-4xl font-bold text-gray-900">{{ league.name }}</h1>
                            <span class="px-3 py-1 rounded-full text-sm font-semibold" :class="{
                                'bg-green-100 text-green-700': league.status === 'active',
                                'bg-blue-100 text-blue-700': league.status === 'upcoming',
                                'bg-gray-100 text-gray-700': league.status === 'completed'
                            }">
                                {{ league.status }}
                            </span>
                        </div>
                        <p class="text-gray-600 mb-4">{{ league.description }}</p>

                        <!-- League Info -->
                        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                            <div>
                                <span class="text-gray-500">Format:</span>
                                <span class="ml-2 font-semibold">{{ league.format }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Teams:</span>
                                <span class="ml-2 font-semibold">{{ league.currentTeams }}/{{ league.maxTeams }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Location:</span>
                                <span class="ml-2 font-semibold">{{ league.location }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Organizer:</span>
                                <span class="ml-2 font-semibold">{{ league.organizerName }}</span>
                            </div>
                        </div>
                    </div>

                    <!-- Action Buttons -->
                    <div class="flex flex-col gap-2 ml-4">
                        <!-- Organizer Actions -->
                        <template v-if="isOrganizer">
                            <button @click="createGame"
                                class="px-4 py-2 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 whitespace-nowrap">
                                + New Game
                            </button>
                            <button @click="editLeague"
                                class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                Edit League
                            </button>
                        </template>

                        <!-- Member Actions -->
                        <template v-else-if="isMember">
                            <button @click="leaveLeague"
                                class="px-4 py-2 text-red-600 border-2 border-red-600 rounded-lg font-semibold hover:bg-red-50">
                                Leave League
                            </button>
                        </template>

                        <!-- Non-member Actions -->
                        <template v-else>
                            <button v-if="league.currentTeams < league.maxTeams" @click="joinLeague"
                                class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                Join League
                            </button>
                        </template>
                    </div>
                </div>
            </div>

            <!-- Tabs -->
            <div class="bg-white rounded-xl shadow-sm border border-gray-200">
                <div class="border-b border-gray-200">
                    <nav class="flex -mb-px">
                        <button @click="activeTab = 'overview'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'overview'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Overview
                        </button>
                        <button @click="activeTab = 'standings'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'standings'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Standings
                        </button>
                        <button @click="activeTab = 'games'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'games'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Games ({{ games.length }})
                        </button>
                        <button @click="activeTab = 'members'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'members'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Members ({{ members.length }})
                        </button>
                    </nav>
                </div>

                <!-- Tab Content -->
                <div class="p-6">

                    <!-- Overview Tab -->
                    <div v-if="activeTab === 'overview'">
                        <div class="grid grid-cols-1 md:grid-cols-3 gap-6">

                            <!-- Quick Stats -->
                            <div class="bg-blue-50 rounded-lg p-6">
                                <h3 class="text-sm font-semibold text-gray-700 mb-2">Total Games</h3>
                                <p class="text-4xl font-bold text-blue-600">{{ games.length }}</p>
                            </div>

                            <div class="bg-green-50 rounded-lg p-6">
                                <h3 class="text-sm font-semibold text-gray-700 mb-2">Active Players</h3>
                                <p class="text-4xl font-bold text-green-600">{{ members.length }}</p>
                            </div>

                            <div class="bg-purple-50 rounded-lg p-6">
                                <h3 class="text-sm font-semibold text-gray-700 mb-2">Start Date</h3>
                                <p class="text-lg font-bold text-purple-600">
                                    {{ new Date(league.startDate).toLocaleDateString() }}
                                </p>
                            </div>

                        </div>

                        <!-- Recent Games -->
                        <div v-if="games.length > 0" class="mt-6">
                            <h3 class="text-xl font-bold text-gray-900 mb-4">Recent Games</h3>
                            <div class="space-y-3">
                                <div v-for="game in games.slice(0, 5)" :key="game.id"
                                    class="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer"
                                    @click="router.push(`/games/${game.id}`)">
                                    <div>
                                        <p class="font-semibold">{{ game.team1 }} vs {{ game.team2 }}</p>
                                        <p class="text-sm text-gray-600">{{ new
                                            Date(game.createdAt).toLocaleDateString() }}</p>
                                    </div>
                                    <div class="text-right">
                                        <p class="font-bold text-lg">{{ game.team1Score }} - {{ game.team2Score }}</p>
                                        <span class="text-xs px-2 py-1 rounded" :class="{
                                            'bg-green-100 text-green-700': game.status === 'completed',
                                            'bg-yellow-100 text-yellow-700': game.status === 'active'
                                        }">
                                            {{ game.status }}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Standings Tab -->
                    <div v-if="activeTab === 'standings'">
                        <div v-if="standings.length === 0" class="text-center py-12 text-gray-500">
                            No standings yet. Games need to be completed.
                        </div>
                        <div v-else class="overflow-x-auto">
                            <table class="w-full">
                                <thead class="bg-gray-50">
                                    <tr>
                                        <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Rank</th>
                                        <th class="px-4 py-3 text-left text-sm font-semibold text-gray-700">Team</th>
                                        <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">Wins</th>
                                        <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">Losses
                                        </th>
                                        <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">Win %</th>
                                        <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">Points
                                        </th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-gray-200">
                                    <tr v-for="(team, index) in standings" :key="team.id" class="hover:bg-gray-50">
                                        <td class="px-4 py-3 text-sm font-bold">{{ index + 1 }}</td>
                                        <td class="px-4 py-3 text-sm font-semibold">{{ team.name }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.wins }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.losses }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.winPercentage }}%</td>
                                        <td class="px-4 py-3 text-sm text-center font-bold">{{ team.points }}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- Games Tab -->
                    <div v-if="activeTab === 'games'">
                        <div v-if="games.length === 0" class="text-center py-12 text-gray-500">
                            No games yet.
                            <button v-if="isOrganizer" @click="createGame"
                                class="block mx-auto mt-4 px-6 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                Create First Game
                            </button>
                        </div>
                        <div v-else class="space-y-3">
                            <div v-for="game in games" :key="game.id"
                                class="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer"
                                @click="router.push(`/games/${game.id}`)">
                                <div>
                                    <p class="font-semibold">{{ game.team1 }} vs {{ game.team2 }}</p>
                                    <p class="text-sm text-gray-600">{{ new Date(game.createdAt).toLocaleDateString() }}
                                    </p>
                                </div>
                                <div class="text-right">
                                    <p class="font-bold text-lg">{{ game.team1Score }} - {{ game.team2Score }}</p>
                                    <span class="text-xs px-2 py-1 rounded" :class="{
                                        'bg-green-100 text-green-700': game.status === 'completed',
                                        'bg-yellow-100 text-yellow-700': game.status === 'active',
                                        'bg-gray-100 text-gray-700': game.status === 'pending'
                                    }">
                                        {{ game.status }}
                                    </span>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Members Tab -->
                    <div v-if="activeTab === 'members'">
                        <div v-if="members.length === 0" class="text-center py-12 text-gray-500">
                            No members yet.
                        </div>
                        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div v-for="member in members" :key="member.playerId" class="p-4 bg-gray-50 rounded-lg">
                                <p class="font-semibold text-gray-900">{{ member.playerName }}</p>
                                <p class="text-sm text-gray-600">Joined {{ new
                                    Date(member.joinedAt).toLocaleDateString() }}</p>
                                <span v-if="member.playerId === league.organizerId"
                                    class="inline-block mt-2 px-2 py-1 bg-purple-100 text-purple-700 text-xs font-semibold rounded">
                                    Organizer
                                </span>
                            </div>
                        </div>
                    </div>

                </div>
            </div>

        </div>

    </div>
</template>