import { defineStore } from 'pinia'

export const useUserStore = defineStore('user', {
    state: () => ({
        user: null as string | null,
        role: null as 'player' | 'organizer' | null
    }),

    actions: {
        setUser(name: string, role: 'player' | 'organizer') {
            this.user = name
            this.role = role
        },

        logout() {
            this.user = null
            this.role = null
        }
    }
})