<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'

const router = useRouter()
const api = useApi()
const auth = useAuth()

const myPlayer = ref<any>(null)
const teams = ref<any[]>([])
const teamMembers = ref<Record<number, any[]>>({})

const isLoading = ref(true)
const error = ref<string | null>(null)
const createTeamName = ref('')
const showCreateForm = ref(false)

const myTeamId = computed(() => myPlayer.value?.team_id ?? null)
const myTeam = computed(() => myTeamId.value ? teams.value.find(t => t.id === myTeamId.value) : null)
const myTeamMembers = computed(() => myTeamId.value ? (teamMembers.value[myTeamId.value] || []) : [])

async function fetchData() {
    isLoading.value = true
    error.value = null
    try {
        const [playerRes, teamsRes] = await Promise.all([
            api.fetch<any>('/players/me'),
            api.fetch<any[]>('/teams')
        ])
        myPlayer.value = playerRes
        teams.value = teamsRes

        // Load members for teams with 1-2 members (skip empty)
        const memberLoads = teamsRes
            .filter((t: any) => t.member_count > 0)
            .map((t: any) =>
                api.fetch<any[]>(`/teams/${t.id}/members`).then(members => {
                    teamMembers.value[t.id] = members
                }).catch(() => {})
            )
        await Promise.all(memberLoads)
    } catch (err: any) {
        error.value = err.data?.error || err.message
    } finally {
        isLoading.value = false
    }
}

async function createTeam() {
    if (!createTeamName.value.trim()) return
    try {
        await api.fetch('/teams', {
            method: 'POST',
            body: { name: createTeamName.value.trim() }
        })
        createTeamName.value = ''
        showCreateForm.value = false
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

async function joinTeam(teamId: number) {
    try {
        await api.fetch(`/teams/${teamId}/join`, { method: 'POST' })
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

async function leaveTeam() {
    if (!confirm('Are you sure you want to leave your team?')) return
    try {
        await api.fetch('/teams/leave', { method: 'POST' })
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

const searchQuery = ref('')

const otherTeams = computed(() => {
    const query = searchQuery.value.trim().toLowerCase()
    const all = teams.value.filter(t => t.id !== myTeamId.value)
    if (!query) return all
    return all.filter(t => {
        if (t.name.toLowerCase().includes(query)) return true
        return (teamMembers.value[t.id] || []).some((m: any) =>
            m.name.toLowerCase().includes(query)
        )
    })
})

onMounted(fetchData)
</script>

<template>
    <div class="max-w-4xl mx-auto py-8 px-4 space-y-6">

        <div class="flex items-center justify-between">
            <h1 class="text-3xl font-bold text-gray-900">Teams</h1>
            <button v-if="!myTeamId && !showCreateForm" @click="showCreateForm = true"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                Create Team
            </button>
        </div>

        <div v-if="isLoading" class="text-center py-12 text-gray-500">Loading...</div>
        <div v-else-if="error" class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-600">{{ error }}</div>

        <template v-else>

            <!-- Create Team Form -->
            <div v-if="showCreateForm" class="bg-white rounded-xl border shadow-sm p-6">
                <h2 class="text-lg font-semibold mb-4">Create a New Team</h2>
                <div class="flex gap-3">
                    <input v-model="createTeamName" type="text" placeholder="Team name"
                        class="flex-1 px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500"
                        @keydown.enter="createTeam" />
                    <button @click="createTeam"
                        class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                        Create
                    </button>
                    <button @click="showCreateForm = false"
                        class="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg font-semibold hover:bg-gray-300">
                        Cancel
                    </button>
                </div>
            </div>

            <!-- Your Team -->
            <div class="bg-white rounded-xl border shadow-sm p-6">
                <h2 class="text-lg font-semibold mb-4">Your Team</h2>

                <div v-if="!myTeamId" class="text-center py-6 text-gray-500">
                    <p class="mb-4">You're not on a team yet.</p>
                    <p class="text-sm">Create a new team or join an existing one below.</p>
                </div>

                <div v-else>
                    <div class="flex items-center justify-between mb-4">
                        <div>
                            <p class="text-2xl font-bold text-gray-900">{{ myTeam?.name }}</p>
                            <p class="text-sm text-gray-500">{{ myTeamMembers.length }}/2 players</p>
                        </div>
                        <button @click="leaveTeam"
                            class="px-4 py-2 text-red-600 border-2 border-red-600 rounded-lg font-semibold hover:bg-red-50">
                            Leave Team
                        </button>
                    </div>

                    <div class="space-y-2">
                        <div v-for="member in myTeamMembers" :key="member.player_id"
                            class="flex items-center gap-3 p-3 bg-gray-50 rounded-lg">
                            <div class="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold text-sm">
                                {{ member.name[0]?.toUpperCase() }}
                            </div>
                            <div>
                                <p class="font-semibold text-gray-900">{{ member.name }}</p>
                            </div>
                            <span v-if="member.user_id === auth.currentUser.value?.id"
                                class="ml-auto text-xs font-semibold px-2 py-1 bg-blue-100 text-blue-700 rounded">
                                You
                            </span>
                        </div>

                        <!-- Empty slot -->
                        <div v-if="myTeamMembers.length < 2"
                            class="flex items-center gap-3 p-3 border-2 border-dashed border-gray-300 rounded-lg text-gray-400">
                            <div class="w-8 h-8 rounded-full border-2 border-dashed border-gray-300 flex items-center justify-center">
                                +
                            </div>
                            <p class="text-sm">Waiting for a partner to join...</p>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Browse Teams -->
            <div class="bg-white rounded-xl border shadow-sm p-6">
                <div class="flex items-center justify-between mb-4">
                    <h2 class="text-lg font-semibold">
                        {{ myTeamId ? 'All Teams' : 'Find a Team to Join' }}
                    </h2>
                    <span class="text-sm text-gray-500">{{ otherTeams.length }} team{{ otherTeams.length !== 1 ? 's' : '' }}</span>
                </div>

                <input v-model="searchQuery" type="text" placeholder="Search by team name or player name..."
                    class="w-full px-4 py-2 mb-4 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />

                <div v-if="otherTeams.length === 0 && !searchQuery" class="text-center py-6 text-gray-500">
                    No other teams yet.
                </div>
                <div v-else-if="otherTeams.length === 0" class="text-center py-6 text-gray-500">
                    No teams match "{{ searchQuery }}".
                </div>

                <div v-else class="space-y-3">
                    <div v-for="team in otherTeams" :key="team.id"
                        class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div>
                            <p class="font-semibold text-gray-900">{{ team.name }}</p>
                            <div class="flex gap-2 mt-1">
                                <span v-for="member in (teamMembers[team.id] || [])" :key="member.player_id"
                                    class="text-xs text-gray-600 bg-white px-2 py-1 rounded border">
                                    {{ member.name }}
                                </span>
                                <span v-if="(teamMembers[team.id] || []).length === 0"
                                    class="text-xs text-gray-400">No members yet</span>
                            </div>
                        </div>

                        <div class="flex items-center gap-2">
                            <span class="text-sm text-gray-500">{{ team.member_count }}/2</span>
                            <button v-if="!myTeamId && team.member_count < 2" @click="joinTeam(team.id)"
                                class="px-4 py-2 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 text-sm">
                                Join
                            </button>
                            <span v-else-if="team.member_count >= 2"
                                class="px-3 py-1 bg-gray-200 text-gray-500 rounded-lg text-sm">
                                Full
                            </span>
                        </div>
                    </div>
                </div>
            </div>

        </template>

    </div>
</template>
