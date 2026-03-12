import { defineNuxtRouteMiddleware, navigateTo } from "nuxt/app"

function decodeTokenPayload(token: string): { exp: number } | null {
    try {
        const base64url = token.split('.')[1]
        if (!base64url) return null
        const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
        return JSON.parse(atob(base64))
    } catch {
        return null
    }
}

// Routes that can be viewed without logging in
const protectedRoutes = [
    '/leagues/create',
    '/leagues/.*/edit',
    '/games/new',
]

function requiresAuth(path: string): boolean {
    return protectedRoutes.some(pattern => new RegExp(`^${pattern}$`).test(path))
}

export default defineNuxtRouteMiddleware((to) => {
    // Only runs on client — localStorage isn't available server-side
    if (import.meta.server) return

    const isAuthPage = to.path.startsWith('/auth/')
    const token = localStorage.getItem('authToken')

    // Redirect logged-in users away from auth pages
    if (isAuthPage && token) {
        const payload = decodeTokenPayload(token)
        if (payload && payload.exp * 1000 > Date.now()) {
            return navigateTo('/')
        }
    }

    // Allow auth pages and public browsing routes through
    if (isAuthPage || !requiresAuth(to.path)) return

    // Require auth for protected action pages
    if (!token) {
        return navigateTo('/auth/login')
    }

    // Check token expiry
    const payload = decodeTokenPayload(token)
    if (!payload || payload.exp * 1000 < Date.now()) {
        localStorage.removeItem('authToken')
        localStorage.removeItem('user')
        return navigateTo('/auth/login')
    }
})
