import { ref, computed } from 'vue'
import type { User, UserRole } from '~/types/user'
import { Permissions } from '~/types/user'

const currentUser = ref<User | null>(null)
const isAuthenticated = ref(false)

export const useAuth = () => {

    function setUser(user: User) {
        currentUser.value = user
        isAuthenticated.value = true
        localStorage.setItem('user', JSON.stringify(user))
    }

    function clearUser() {
        currentUser.value = null
        isAuthenticated.value = false
        localStorage.removeItem('authToken')
        localStorage.removeItem('user')
    }

    function loadUserFromStorage() {
        const authToken = localStorage.getItem('authToken')
        const userJson = localStorage.getItem('user')

        if (authToken && userJson) {
            try {
                const user = JSON.parse(userJson) as User
                currentUser.value = user
                isAuthenticated.value = true
            } catch (error) {
                console.error('Failed to parse user from localStorage', error)
                clearUser()
            }
        }
    }

    // Role checks
    const isPlayer = computed(() =>
        currentUser.value ? Permissions.isPlayer(currentUser.value) : false
    )

    const isOrganizer = computed(() =>
        currentUser.value ? Permissions.isOrganizer(currentUser.value) : false
    )

    const hasRole = (role: UserRole) => {
        return currentUser.value ? Permissions.hasRole(currentUser.value, role) : false
    }

    // Permission checks
    const canCreateLeague = computed(() =>
        currentUser.value ? Permissions.canCreateLeague(currentUser.value) : false
    )

    const canEditLeague = (leagueOwnerId: string) => {
        return currentUser.value ? Permissions.canEditLeague(currentUser.value, leagueOwnerId) : false
    }

    const canEditGameResults = (leagueOwnerId: string) => {
        return currentUser.value ? Permissions.canEditGameResults(currentUser.value, leagueOwnerId) : false
    }

    const canManageLeagueMembers = (leagueOwnerId: string) => {
        return currentUser.value ? Permissions.canManageLeagueMembers(currentUser.value, leagueOwnerId) : false
    }

    return {
        // State
        currentUser,
        isAuthenticated,

        // Role checks
        isPlayer,
        isOrganizer,
        hasRole,

        // Permission checks
        canCreateLeague,
        canEditLeague,
        canEditGameResults,
        canManageLeagueMembers,

        // Actions
        setUser,
        clearUser,
        loadUserFromStorage
    }
}