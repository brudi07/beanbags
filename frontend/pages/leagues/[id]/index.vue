<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '~/composables/useApi'
import { useAuth } from '~/composables/useAuth'
import { useToast } from '~/composables/useToast'
import { useConfirm } from '~/composables/useConfirm'
import type { League, LeagueSchedule, LeagueGame } from '~/types/league'

const route = useRoute()
const router = useRouter()
const auth = useAuth()
const api = useApi()
const toast = useToast()
const confirm = useConfirm().confirm

const leagueId = route.params.id as string

const league = ref<League | null>(null)
const members = ref<any[]>([])
const schedule = ref<LeagueSchedule[]>([])
const standings = ref<any[]>([])
const myPlayer = ref<any>(null)
const myTeamMembers = ref<Record<number, any[]>>({})
const showJoinPanel = ref(false)

const isLoading = ref(true)
const error = ref<string | null>(null)
const activeTab = ref<'schedule' | 'standings' | 'members'>('schedule')

// Date navigation
const selectedDateIndex = ref(0)
const availableDates = computed(() => schedule.value.map(s => s.date).sort())
const selectedDate = computed(() => availableDates.value[selectedDateIndex.value] || '')
const gamesForSelectedDate = computed(() =>
    schedule.value.find(s => s.date === selectedDate.value)?.games || []
)

const canGoToPreviousDate = computed(() => selectedDateIndex.value > 0)
const canGoToNextDate = computed(() => selectedDateIndex.value < availableDates.value.length - 1)

const isOrganizer = computed(() =>
    league.value && auth.currentUser.value &&
    league.value.organizer_id === auth.currentUser.value.id
)

const isMember = computed(() =>
    myPlayer.value != null && members.value.some(m => m.player_id === myPlayer.value?.id)
)

const memberTeams = computed(() => {
    const teams: Record<number, { teamId: number, teamName: string, players: any[] }> = {}
    for (const m of members.value) {
        if (m.team_id) {
            if (!teams[m.team_id]) teams[m.team_id] = { teamId: m.team_id, teamName: m.team_name, players: [] }
            teams[m.team_id]!.players.push(m)
        }
    }
    return Object.values(teams)
})

function canStartGame(game: LeagueGame): boolean {
    if (!auth.currentUser.value) return false

    // Check if current user is one of the players in this game
    return game.team1_player_ids.includes(auth.currentUser.value.id) ||
        game.team2_player_ids.includes(auth.currentUser.value.id)
}

function previousDate() {
    if (canGoToPreviousDate.value) {
        selectedDateIndex.value--
    }
}

function nextDate() {
    if (canGoToNextDate.value) {
        selectedDateIndex.value++
    }
}

function goToToday() {
    if (availableDates.value.length === 0) return

    const today: string = new Date().toISOString().split('T')[0]!
    const todayIndex = availableDates.value.findIndex(date => date === today)

    if (todayIndex !== -1) {
        selectedDateIndex.value = todayIndex
    } else {
        // Find closest future date
        const futureIndex = availableDates.value.findIndex(date => date > today)
        selectedDateIndex.value = futureIndex !== -1 ? futureIndex : Math.max(0, availableDates.value.length - 1)
    }
}

async function fetchLeagueData() {
    isLoading.value = true
    error.value = null

    try {
        // ✅ $fetch returns parsed data directly
        const [leagueRes, membersRes, scheduleRes, standingsRes] = await Promise.all([
            api.fetch<League>(`/leagues/${leagueId}`),
            api.fetch<any[]>(`/leagues/${leagueId}/members`),
            api.fetch<LeagueSchedule[]>(`/leagues/${leagueId}/schedule`),
            api.fetch<any[]>(`/leagues/${leagueId}/standings`)
        ])

        league.value = leagueRes
        members.value = membersRes
        schedule.value = scheduleRes
        standings.value = standingsRes

        goToToday()

        try {
            myPlayer.value = await api.fetch<any>('/players/me')
            // Fetch members for all the user's teams so the join panel can show them
            const memberLoads = (myPlayer.value?.teams ?? []).map((t: any) =>
                api.fetch<any[]>(`/teams/${t.id}/members`).then((m: any[]) => {
                    myTeamMembers.value[t.id] = m
                }).catch(() => {})
            )
            await Promise.all(memberLoads)
        } catch {} // not critical — fails for non-players (organizer-only accounts)
    } catch (err: any) {
        error.value = err.data?.error || err.message
    } finally {
        isLoading.value = false
    }
}

async function startGame(game: LeagueGame) {
    if (!canStartGame(game)) {
        toast.error('You are not a player in this game')
        return
    }

    if (game.status === 'completed') {
        // View completed game
        router.push(`/games/${game.game_id}`)
        return
    }

    if (game.game_id) {
        // Continue in-progress game
        router.push(`/games/${game.game_id}`)
    } else {
        // Create new game - pass player names so games/new.vue can pre-fill
        const team1Names = game.team1_player_ids.map(id =>
            members.value.find(m => m.player_id === id)?.player_name || ''
        )
        const team2Names = game.team2_player_ids.map(id =>
            members.value.find(m => m.player_id === id)?.player_name || ''
        )
        const params = new URLSearchParams({
            leagueId,
            leagueGameId: String(game.id),
            format: String(league.value!.format),
            bestOf: String(league.value!.games_per_match),
            t1p1: team1Names[0] || '',
            t2p1: team2Names[0] || '',
        })
        if (team1Names[1]) params.set('t1p2', team1Names[1])
        if (team2Names[1]) params.set('t2p2', team2Names[1])
        router.push(`/games/new?${params.toString()}`)
    }
}

const rescheduleGameId = ref<number | null>(null)
const newScheduledDate = ref('')

function openReschedule(gameId: number, currentDate: string) {
    rescheduleGameId.value = gameId
    newScheduledDate.value = currentDate
}

function closeReschedule() {
    rescheduleGameId.value = null
    newScheduledDate.value = ''
}

async function confirmReschedule() {
    if (!rescheduleGameId.value || !newScheduledDate.value) return

    try {
        // ✅ No Response cast, no .ok check, no .json()
        await api.fetch(`/leagues/${leagueId}/games/${rescheduleGameId.value}/reschedule`, {
            method: 'PATCH',
            body: { scheduledDate: newScheduledDate.value }
        })

        closeReschedule()
        await fetchLeagueData()
    } catch (err: any) {
        toast.error(err.data?.error || err.message)
    }
}

async function joinLeague(teamId?: number) {
    try {
        await api.fetch(`/leagues/${leagueId}/join`, {
            method: 'POST',
            body: teamId ? { team_id: teamId } : undefined
        })
        showJoinPanel.value = false
        await fetchLeagueData()
    } catch (err: any) {
        toast.error(err.data?.error || err.message)
    }
}

async function leaveLeague() {
    if (!await confirm('Are you sure you want to leave this league?')) return

    try {
        // ✅ No Response cast needed
        await api.fetch(`/leagues/${leagueId}/leave`, {
            method: 'POST',
        })

        router.push('/leagues')
    } catch (err: any) {
        toast.error(err.data?.error || err.message)
    }
}

function editLeague() {
    router.push(`/leagues/${leagueId}/edit`)
}

async function scheduleGames() {
    if (!await confirm('Generate the league schedule? This cannot be undone.', {
        title: 'Generate Schedule',
        confirmLabel: 'Generate',
    })) {
        return
    }

    try {
        const result = await api.fetch<{
            message: string
            total_matches: number
            total_games: number
        }>(`/leagues/${leagueId}/generate-schedule`, {
            method: 'POST'
        })

        toast.success(`Schedule created! ${result.total_matches} matches, ${result.total_games} games`)

        await fetchLeagueData()
    } catch (err: any) {
        toast.error(err.data?.error || err.message)
    }
}

function getPlayerNames(playerIds: number[]): string {
    return playerIds
        .map(id => members.value.find(m => m.player_id === id)?.player_name || 'Unknown')
        .join(' & ')
}

function getGameStatusColor(status: string) {
    switch (status) {
        case 'completed': return 'bg-green-100 text-green-700'
        case 'in_progress': return 'bg-yellow-100 text-yellow-700'
        case 'scheduled': return 'bg-gray-100 text-gray-700'
        default: return 'bg-gray-100 text-gray-700'
    }
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
                        <div class="grid grid-cols-2 md:grid-cols-5 gap-4 text-sm">
                            <div>
                                <span class="text-gray-500">Format:</span>
                                <span class="ml-2 font-semibold">{{ league.format }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Match:</span>
                                <span class="ml-2 font-semibold">Best of {{ league.games_per_match }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Teams:</span>
                                <span class="ml-2 font-semibold">{{ league.current_teams }}/{{ league.max_teams
                                }}</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Duration:</span>
                                <span class="ml-2 font-semibold">{{ league.weeks_of_play }} weeks</span>
                            </div>
                            <div>
                                <span class="text-gray-500">Location:</span>
                                <span class="ml-2 font-semibold">{{ league.location }}</span>
                            </div>
                        </div>
                    </div>

                    <!-- Action Buttons -->
                    <div class="flex flex-col gap-2 ml-4">
                        <template v-if="isOrganizer">
                            <button @click="editLeague"
                                class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                Edit League
                            </button>

                           <button v-if="isOrganizer && schedule.length === 0 && league.current_teams >= 2" @click="scheduleGames"
                                class="px-4 py-2 bg-purple-600 text-white rounded-lg font-semibold hover:bg-purple-700">
                                Schedule Games
                            </button>
                            <p v-if="isOrganizer && schedule.length === 0 && league.current_teams < 2"
                                class="text-xs text-gray-500 text-center">
                                Need at least 2 teams to schedule
                            </p>

                            <!-- Allow organizer to join as player if not already a member -->
                            <button v-if="!isMember && league.current_teams < league.max_teams" @click="league.format === '2v2' ? showJoinPanel = true : joinLeague()"
                                class="px-4 py-2 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700">
                                {{ league.format === '2v2' ? 'Join as Team' : 'Join as Player' }}
                            </button>

                            <!-- Allow organizer to leave if they joined as player/team -->
                            <button v-else-if="isMember" @click="leaveLeague"
                                class="px-4 py-2 text-red-600 border-2 border-red-600 rounded-lg font-semibold hover:bg-red-50">
                                {{ league.format === '2v2' ? 'Leave as Team' : 'Leave as Player' }}
                            </button>
                        </template>

                        <template v-else-if="isMember">
                            <button @click="leaveLeague"
                                class="px-4 py-2 text-red-600 border-2 border-red-600 rounded-lg font-semibold hover:bg-red-50">
                                {{ league.format === '2v2' ? 'Leave as Team' : 'Leave League' }}
                            </button>
                        </template>

                        <template v-else>
                            <div v-if="league.current_teams < league.max_teams">
                                <!-- 1v1: join directly -->
                                <button v-if="league.format === '1v1'" @click="joinLeague()"
                                    class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                    Join League
                                </button>

                                <!-- 2v2: open team confirmation panel -->
                                <div v-else>
                                    <button @click="showJoinPanel = !showJoinPanel"
                                        class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                        Join as Team
                                    </button>
                                </div>
                            </div>
                        </template>
                    </div>

                </div>
            </div>

            <!-- Tabs -->
            <div class="bg-white rounded-xl shadow-sm border border-gray-200">
                <div class="border-b border-gray-200">
                    <nav class="flex -mb-px">
                        <button @click="activeTab = 'schedule'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'schedule'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Schedule
                        </button>
                        <button @click="activeTab = 'standings'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'standings'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Standings
                        </button>
                        <button @click="activeTab = 'members'"
                            class="px-6 py-4 font-semibold border-b-2 transition-colors" :class="activeTab === 'members'
                                ? 'border-blue-600 text-blue-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700'">
                            Members ({{ league.format === '2v2' ? memberTeams.length : members.length }})
                        </button>
                    </nav>
                </div>

                <!-- Tab Content -->
                <div class="p-6">

                    <!-- Schedule Tab -->
                    <div v-if="activeTab === 'schedule'">

                        <!-- Date Navigation -->
                        <div v-if="availableDates.length > 0" class="mb-6">
                            <div class="flex items-center justify-between bg-gray-50 rounded-lg p-4">

                                <!-- Previous Button -->
                                <button @click="previousDate" :disabled="!canGoToPreviousDate"
                                    class="p-2 rounded-lg transition-colors" :class="canGoToPreviousDate
                                        ? 'hover:bg-gray-200 text-gray-700'
                                        : 'text-gray-300 cursor-not-allowed'">
                                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M15 19l-7-7 7-7" />
                                    </svg>
                                </button>

                                <!-- Current Date Display -->
                                <div class="text-center">
                                    <p class="text-2xl font-bold text-gray-900">
                                        {{ selectedDate ? new Date(selectedDate + 'T00:00:00').toLocaleDateString('en-US', {
                                            weekday: 'long',
                                            month: 'long',
                                            day: 'numeric',
                                            year: 'numeric'
                                        }) : 'No date selected' }}
                                    </p>
                                    <p class="text-sm text-gray-600 mt-1">
                                        {{ gamesForSelectedDate.length }} game{{ gamesForSelectedDate.length !== 1 ? 's'
                                            : '' }}
                                        scheduled
                                    </p>
                                </div>

                                <!-- Next Button -->
                                <button @click="nextDate" :disabled="!canGoToNextDate"
                                    class="p-2 rounded-lg transition-colors" :class="canGoToNextDate
                                        ? 'hover:bg-gray-200 text-gray-700'
                                        : 'text-gray-300 cursor-not-allowed'">
                                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                            d="M9 5l7 7-7 7" />
                                    </svg>
                                </button>
                            </div>

                            <!-- Jump to Today -->
                            <div class="text-center mt-2">
                                <button @click="goToToday"
                                    class="text-sm text-blue-600 hover:text-blue-700 font-medium">
                                    Jump to Today
                                </button>
                            </div>
                        </div>

                        <!-- Games for Selected Date -->
                        <div v-if="gamesForSelectedDate.length === 0" class="text-center py-12 text-gray-500">
                            No games scheduled for this date
                        </div>

                        <div v-else class="space-y-4">
                            <div v-for="game in gamesForSelectedDate" :key="game.id"
                                class="bg-gray-50 rounded-lg p-6 hover:bg-gray-100 transition-colors">
                                <div class="flex items-center justify-between">

                                    <!-- Game Info -->
                                    <div class="flex-1">
                                        <div class="flex items-center gap-3 mb-2">
                                            <span class="font-semibold text-gray-700">Match {{ game.match_number
                                            }}</span>
                                            <span class="text-gray-400">•</span>
                                            <span class="text-sm text-gray-600">Game {{ game.game_number }} of {{
                                                league.games_per_match }}</span>
                                            <span class="px-2 py-1 rounded text-xs font-semibold"
                                                :class="getGameStatusColor(game.status)">
                                                {{ game.status.replace('_', ' ') }}
                                            </span>
                                        </div>

                                        <div class="text-lg font-bold text-gray-900 mb-1">
                                            {{ getPlayerNames(game.team1_player_ids) }}
                                            <span class="text-gray-400 mx-2">vs</span>
                                            {{ getPlayerNames(game.team2_player_ids) }}
                                        </div>

                                        <div v-if="game.status === 'completed' && game.team1_score != null" class="mt-2 flex items-center gap-3">
                                            <span class="text-2xl font-bold" :class="game.winning_team === 1 ? 'text-green-600' : 'text-gray-500'">
                                                {{ game.team1_score }}
                                            </span>
                                            <span class="text-gray-400 font-semibold">–</span>
                                            <span class="text-2xl font-bold" :class="game.winning_team === 2 ? 'text-green-600' : 'text-gray-500'">
                                                {{ game.team2_score }}
                                            </span>
                                            <span class="text-sm text-green-600 font-semibold ml-1">
                                                ({{ game.winning_team === 1 ? getPlayerNames(game.team1_player_ids) : getPlayerNames(game.team2_player_ids) }} wins)
                                            </span>
                                        </div>
                                    </div>

                                    <!-- Actions -->
                                    <div class="flex items-center gap-2 ml-4">

                                        <!-- Organizer: Reschedule -->
                                        <button v-if="isOrganizer && game.status === 'scheduled'"
                                            @click="openReschedule(game.id, game.scheduled_date)"
                                            class="px-3 py-2 text-sm text-gray-700 border border-gray-300 rounded-lg hover:bg-gray-50">
                                            Reschedule
                                        </button>

                                        <!-- Player: Start/View Game -->
                                        <button v-if="canStartGame(game)" @click="startGame(game)"
                                            class="px-4 py-2 rounded-lg font-semibold text-white transition-colors"
                                            :class="{
                                                'bg-green-600 hover:bg-green-700': game.status === 'scheduled',
                                                'bg-yellow-600 hover:bg-yellow-700': game.status === 'in_progress',
                                                'bg-blue-600 hover:bg-blue-700': game.status === 'completed'
                                            }">
                                            {{ game.status === 'completed' ? 'View Game' : game.status === 'in_progress'
                                                ?
                                                'Continue' : 'Start Game' }}
                                        </button>

                                        <!-- Non-player: View only if completed -->
                                        <button v-else-if="game.status === 'completed'"
                                            @click="router.push(`/games/${game.game_id}`)"
                                            class="px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                                            View Game
                                        </button>
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
                                        <th class="px-4 py-3 text-center text-sm font-semibold text-gray-700">+/-</th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-gray-200">
                                    <tr v-for="(team, index) in standings" :key="team.team_id" class="hover:bg-gray-50">
                                        <td class="px-4 py-3 text-sm font-bold">{{ index + 1 }}</td>
                                        <td class="px-4 py-3 text-sm font-semibold">{{ team.team_name }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.wins }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.losses }}</td>
                                        <td class="px-4 py-3 text-sm text-center">{{ team.win_percentage }}%</td>
                                        <td class="px-4 py-3 text-sm text-center font-bold">{{ team.point_diff }}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- Members Tab -->
                    <div v-if="activeTab === 'members'">
                        <div v-if="members.length === 0" class="text-center py-12 text-gray-500">
                            No members yet.
                        </div>
                        <!-- 2v2: show teams -->
                        <div v-else-if="league.format === '2v2'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div v-for="team in memberTeams" :key="team.teamId" class="p-4 bg-gray-50 rounded-lg">
                                <p class="font-semibold text-gray-900 mb-2">{{ team.teamName }}</p>
                                <div class="space-y-1">
                                    <p v-for="player in team.players" :key="player.player_id"
                                        class="text-sm text-gray-600">{{ player.player_name }}</p>
                                </div>
                                <p class="text-xs text-gray-400 mt-2">Joined {{ new Date(team.players[0].joined_at).toLocaleDateString() }}</p>
                            </div>
                        </div>
                        <!-- 1v1: show individual players -->
                        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                            <div v-for="member in members" :key="member.player_id" class="p-4 bg-gray-50 rounded-lg">
                                <p class="font-semibold text-gray-900">{{ member.player_name }}</p>
                                <p class="text-sm text-gray-600">Joined {{ new Date(member.joined_at).toLocaleDateString() }}</p>
                            </div>
                        </div>
                    </div>

                </div>
            </div>

        </div>

        <!-- 2v2 Team Join Modal -->
        <div v-if="showJoinPanel && league?.format === '2v2'"
            class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
            @click.self="showJoinPanel = false">
            <div class="bg-white rounded-xl shadow-2xl p-6 max-w-md w-full mx-4">
                <h3 class="text-xl font-bold text-gray-900 mb-4">Join as Team</h3>

                <!-- No teams -->
                <div v-if="!myPlayer?.teams?.length" class="text-center py-4">
                    <p class="text-gray-700 mb-4">You need a team to join a 2v2 league.</p>
                    <NuxtLink to="/teams"
                        class="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                        Go to Teams →
                    </NuxtLink>
                </div>

                <!-- Team list -->
                <div v-else class="space-y-3">
                    <p class="text-sm text-gray-600">Select a team to join with:</p>
                    <div v-for="team in myPlayer.teams" :key="team.id" class="p-4 bg-gray-50 rounded-lg border">
                        <p class="font-semibold text-gray-900 mb-2">{{ team.name }}</p>
                        <!-- Team not full -->
                        <div v-if="(myTeamMembers[team.id]?.length ?? 0) < 2">
                            <p class="text-sm text-amber-600 mb-2">
                                Needs 2 players ({{ myTeamMembers[team.id]?.length ?? 0 }}/2)
                            </p>
                            <NuxtLink to="/teams" class="text-sm text-blue-600 font-semibold hover:underline">
                                Invite a partner →
                            </NuxtLink>
                        </div>
                        <!-- Team ready -->
                        <div v-else>
                            <p class="text-sm text-gray-500 mb-3">
                                {{ myTeamMembers[team.id]?.map((m: any) => m.name).join(' & ') }}
                            </p>
                            <button @click="joinLeague(team.id)"
                                class="px-4 py-2 bg-green-600 text-white rounded-lg font-semibold hover:bg-green-700 text-sm">
                                Confirm Join
                            </button>
                        </div>
                    </div>
                </div>

                <button @click="showJoinPanel = false"
                    class="mt-4 w-full px-4 py-2 bg-gray-200 text-gray-700 rounded-lg font-semibold hover:bg-gray-300">
                    Cancel
                </button>
            </div>
        </div>

        <!-- Reschedule Modal -->
        <div v-if="rescheduleGameId" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
            @click.self="closeReschedule">
            <div class="bg-white rounded-xl shadow-2xl p-6 max-w-md w-full mx-4">
                <h3 class="text-xl font-bold text-gray-900 mb-4">Reschedule Game</h3>

                <div class="mb-6">
                    <label for="newDate" class="block text-sm font-semibold text-gray-700 mb-2">
                        New Date
                    </label>
                    <input id="newDate" v-model="newScheduledDate" type="date"
                        class="w-full px-4 py-3 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500" />
                </div>

                <div class="flex gap-3">
                    <button @click="confirmReschedule"
                        class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg font-semibold hover:bg-blue-700">
                        Confirm
                    </button>
                    <button @click="closeReschedule"
                        class="flex-1 px-4 py-2 bg-gray-200 text-gray-700 rounded-lg font-semibold hover:bg-gray-300">
                        Cancel
                    </button>
                </div>
            </div>
        </div>

    </div>
</template>