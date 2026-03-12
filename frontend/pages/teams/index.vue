<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'

const api = useApi()
const auth = useAuth()

const myPlayer = ref<any>(null)
const teams = ref<any[]>([])
const teamMembers = ref<Record<number, any[]>>({})
const publicLeagues = ref<any[]>([])
const teamInterestSearches = ref<Record<number, string>>({})

const isLoading = ref(true)
const error = ref<string | null>(null)
const createTeamName = ref('')
const showCreateForm = ref(false)

const myTeamIds = computed(() => new Set((myPlayer.value?.teams ?? []).map((t: any) => t.id)))

async function fetchData() {
    isLoading.value = true
    error.value = null
    try {
        const requests: Promise<any>[] = [api.fetch<any[]>('/teams')]
        if (auth.isAuthenticated.value) {
            requests.push(api.fetch<any>('/players/me'))
        }
        const [teamsRes, playerRes] = await Promise.all(requests)
        teams.value = teamsRes
        myPlayer.value = playerRes ?? null

        api.fetch<any[]>('/leagues/public').then(r => { publicLeagues.value = r }).catch(() => {})

        // Load members for all teams with members
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

async function leaveTeam(teamId: number) {
    if (!confirm('Are you sure you want to leave this team?')) return
    try {
        await api.fetch('/teams/leave', {
            method: 'POST',
            body: { team_id: teamId }
        })
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

async function addInterest(teamId: number, leagueId: number) {
    try {
        await api.fetch(`/teams/${teamId}/interests`, {
            method: 'POST',
            body: { league_id: leagueId }
        })
        teamInterestSearches.value[teamId] = ''
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

async function removeInterest(teamId: number, leagueId: number) {
    try {
        await api.fetch(`/teams/${teamId}/interests/${leagueId}`, { method: 'DELETE' })
        await fetchData()
    } catch (err: any) {
        alert(err.data?.error || err.message)
    }
}

const searchQuery = ref('')
const fillFilter = ref<'open' | 'full' | 'all'>('open')

const browsableTeams = computed(() => {
    const query = searchQuery.value.trim().toLowerCase()
    return teams.value.filter(t => {
        if (fillFilter.value === 'open' && t.member_count >= 2) return false
        if (fillFilter.value === 'full' && t.member_count < 2) return false
        if (!query) return true
        if (t.name.toLowerCase().includes(query)) return true
        if ((teamMembers.value[t.id] || []).some((m: any) => m.name.toLowerCase().includes(query))) return true
        if ((t.interested_leagues || []).some((l: any) => l.name.toLowerCase().includes(query))) return true
        return false
    })
})

onMounted(fetchData)
</script>

<template>
    <div class="max-w-4xl mx-auto py-8 px-4 space-y-6">

        <div class="flex items-center justify-between">
            <h1 class="text-3xl font-bold text-gray-900">Teams</h1>
            <button v-if="auth.isAuthenticated.value && !showCreateForm" @click="showCreateForm = true"
                class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                Create Team
            </button>
        </div>

        <div v-if="isLoading" class="text-center py-12 text-gray-500">Loading Teams...</div>
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

            <!-- Your Teams -->
            <div v-if="auth.isAuthenticated.value" class="bg-white rounded-xl border shadow-sm p-6">
                <h2 class="text-lg font-semibold mb-4">Your Teams</h2>

                <div v-if="!myPlayer?.teams?.length" class="text-center py-6 text-gray-500">
                    <p class="mb-2">You're not on any teams yet.</p>
                    <p class="text-sm">Create a new team or join an existing one below.</p>
                </div>

                <div v-else class="space-y-4">
                    <div v-for="team in myPlayer.teams" :key="team.id"
                        class="border rounded-lg p-4">
                        <div class="flex items-center justify-between mb-3">
                            <div>
                                <p class="text-xl font-bold text-gray-900">{{ team.name }}</p>
                                <p class="text-sm text-gray-500">{{ team.member_count }}/2 players</p>
                            </div>
                            <button @click="leaveTeam(team.id)"
                                class="px-3 py-1.5 text-sm text-red-600 border border-red-300 rounded-lg font-semibold hover:bg-red-50">
                                Leave
                            </button>
                        </div>

                        <div class="space-y-2">
                            <div v-for="member in (teamMembers[team.id] || [])" :key="member.player_id"
                                class="flex items-center gap-3 p-2 bg-gray-50 rounded-lg">
                                <div class="w-7 h-7 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold text-xs">
                                    {{ member.name[0]?.toUpperCase() }}
                                </div>
                                <p class="font-medium text-gray-900 text-sm">{{ member.name }}</p>
                                <span v-if="member.user_id === auth.currentUser.value?.id"
                                    class="ml-auto text-xs font-semibold px-2 py-0.5 bg-blue-100 text-blue-700 rounded">
                                    You
                                </span>
                            </div>
                            <div v-if="(teamMembers[team.id] || []).length < 2"
                                class="flex items-center gap-3 p-2 border-2 border-dashed border-gray-200 rounded-lg text-gray-400">
                                <div class="w-7 h-7 rounded-full border-2 border-dashed border-gray-300 flex items-center justify-center text-xs">+</div>
                                <p class="text-sm">Waiting for a partner...</p>
                            </div>
                        </div>

                        <div class="mt-4 pt-3 border-t border-gray-100">
                            <p class="text-xs font-semibold text-gray-500 uppercase tracking-wide mb-2">Interested Leagues</p>
                            <div class="flex flex-wrap gap-1 mb-2">
                                <span v-for="interest in (teams.find(t => t.id === team.id)?.interested_leagues ?? [])"
                                    :key="interest.id"
                                    class="inline-flex items-center gap-1 text-xs bg-gray-100 text-gray-700 px-2 py-0.5 rounded-full border border-gray-200">
                                    {{ interest.name }}
                                    <button @click="removeInterest(team.id, interest.id)"
                                        class="text-red-400 hover:text-red-600 leading-none font-bold">&times;</button>
                                </span>
                                <span v-if="!(teams.find(t => t.id === team.id)?.interested_leagues ?? []).length"
                                    class="text-xs text-gray-400">No interests added yet</span>
                            </div>
                            <div class="relative">
                                <input v-model="teamInterestSearches[team.id]"
                                    type="text" placeholder="Search leagues..."
                                    class="w-full text-sm px-2 py-1 rounded border border-gray-300 focus:outline-none focus:ring-1 focus:ring-blue-500" />
                                <ul v-if="teamInterestSearches[team.id]?.trim()"
                                    class="absolute z-10 mt-1 w-full bg-white border border-gray-200 rounded shadow-lg max-h-40 overflow-y-auto">
                                    <li v-for="league in publicLeagues.filter(l =>
                                            l.format === '2v2' &&
                                            l.name.toLowerCase().includes((teamInterestSearches[team.id] ?? '').toLowerCase()) &&
                                            !(teams.find(t => t.id === team.id)?.interested_leagues ?? []).some((i: any) => i.id === l.id)
                                        )" :key="league.id"
                                        @click="addInterest(team.id, league.id)"
                                        class="px-3 py-2 text-sm text-gray-800 hover:bg-blue-50 cursor-pointer">
                                        {{ league.name }}
                                    </li>
                                    <li v-if="!publicLeagues.filter(l =>
                                            l.format === '2v2' &&
                                            l.name.toLowerCase().includes((teamInterestSearches[team.id] ?? '').toLowerCase()) &&
                                            !(teams.find(t => t.id === team.id)?.interested_leagues ?? []).some((i: any) => i.id === l.id)
                                        ).length"
                                        class="px-3 py-2 text-sm text-gray-400">No matching leagues</li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Browse Teams -->
            <div class="bg-white rounded-xl border shadow-sm p-6">
                <div class="flex items-center justify-between mb-4">
                    <h2 class="text-lg font-semibold">Browse Teams</h2>
                    <span class="text-sm text-gray-500">{{ browsableTeams.length }} team{{ browsableTeams.length !== 1 ? 's' : '' }}</span>
                </div>

                <div class="flex gap-2 mb-3">
                    <button v-for="opt in [{ value: 'open', label: 'Open' }, { value: 'full', label: 'Full' }, { value: 'all', label: 'All' }]"
                        :key="opt.value"
                        @click="fillFilter = opt.value as 'open' | 'full' | 'all'"
                        class="px-3 py-1 text-sm rounded-full font-semibold border transition-colors"
                        :class="fillFilter === opt.value
                            ? 'bg-blue-600 text-white border-blue-600'
                            : 'bg-white text-gray-600 border-gray-300 hover:border-blue-400'">
                        {{ opt.label }}
                    </button>
                </div>

                <input v-model="searchQuery" type="text" placeholder="Search by team name, player, or league..."
                    class="w-full px-4 py-2 mb-4 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />

                <div v-if="browsableTeams.length === 0" class="text-center py-6 text-gray-500">
                    {{ searchQuery ? `No teams match "${searchQuery}".` : fillFilter === 'open' ? 'No open teams.' : fillFilter === 'full' ? 'No full teams.' : 'No teams yet.' }}
                </div>

                <div v-else class="space-y-3">
                    <div v-for="team in browsableTeams" :key="team.id"
                        class="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div>
                            <p class="font-semibold text-gray-900">{{ team.name }}</p>
                            <div class="flex gap-2 mt-1 flex-wrap">
                                <span v-for="member in (teamMembers[team.id] || [])" :key="member.player_id"
                                    class="text-xs text-gray-600 bg-white px-2 py-1 rounded border">
                                    {{ member.name }}
                                </span>
                                <span v-if="!(teamMembers[team.id] || []).length"
                                    class="text-xs text-gray-400">No members yet</span>
                            </div>
                            <div v-if="team.interested_leagues?.length > 0" class="flex gap-1 mt-2 flex-wrap">
                                <span v-for="league in team.interested_leagues" :key="league.id"
                                    class="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded-full border border-gray-200">
                                    {{ league.name }}
                                </span>
                            </div>
                        </div>

                        <div class="flex items-center gap-2 ml-4 flex-shrink-0">
                            <span class="text-sm text-gray-500">{{ team.member_count }}/2</span>
                            <span v-if="myTeamIds.has(team.id)"
                                class="px-3 py-1 bg-blue-100 text-blue-700 rounded-lg text-sm font-semibold">
                                Yours
                            </span>
                            <button v-else-if="auth.isAuthenticated.value && team.member_count < 2" @click="joinTeam(team.id)"
                                class="px-4 py-2 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 text-sm">
                                Join
                            </button>
                            <span v-else class="px-3 py-1 bg-gray-200 text-gray-500 rounded-lg text-sm">Full</span>
                        </div>
                    </div>
                </div>
            </div>

        </template>

    </div>
</template>
