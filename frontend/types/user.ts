export type UserRole = 'player' | 'organizer'

export interface User {
    id: string
    username: string
    email: string
    roles: UserRole[]
    createdAt: string
    updatedAt: string
}

export interface AuthResponse {
    token: string
    user: User
}

// Permission helpers
export const Permissions = {
    // Player permissions
    canJoinLeague: (user: User) => user.roles.includes('player') || user.roles.includes('organizer'),
    canPlayGames: (user: User) => user.roles.includes('player') || user.roles.includes('organizer'),
    canViewStats: (user: User) => user.roles.includes('player') || user.roles.includes('organizer'),

    // Organizer permissions
    canCreateLeague: (user: User) => user.roles.includes('organizer'),
    canEditLeague: (user: User, leagueOwnerId: string) => {
        return user.roles.includes('organizer') && user.id === leagueOwnerId
    },
    canEditGameResults: (user: User, leagueOwnerId: string) => {
        return user.roles.includes('organizer') && user.id === leagueOwnerId
    },
    canManageLeagueMembers: (user: User, leagueOwnerId: string) => {
        return user.roles.includes('organizer') && user.id === leagueOwnerId
    },

    // Combined checks
    isPlayer: (user: User) => user.roles.includes('player'),
    isOrganizer: (user: User) => user.roles.includes('organizer'),
    hasRole: (user: User, role: UserRole) => user.roles.includes(role)
}