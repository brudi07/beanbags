export default defineNuxtRouteMiddleware((to) => {
    // Only runs on client — localStorage isn't available server-side
    if (process.server) return

    const isAuthPage = to.path.startsWith('/auth/')

    const token = localStorage.getItem('authToken')

    // Redirect logged-in users away from auth pages
    if (isAuthPage && token) {
        try {
            const payload = JSON.parse(atob(token.split('.')[1]))
            if (payload.exp * 1000 > Date.now()) {
                return navigateTo('/')
            }
        } catch {
            // Invalid token — let them through to the auth page
        }
    }

    // Allow auth pages through for unauthenticated users
    if (isAuthPage) return

    // Require auth for all other pages
    if (!token) {
        return navigateTo('/auth/login')
    }

    // Check token expiry
    try {
        const payload = JSON.parse(atob(token.split('.')[1]))
        if (payload.exp * 1000 < Date.now()) {
            localStorage.removeItem('authToken')
            localStorage.removeItem('user')
            return navigateTo('/auth/login')
        }
    } catch {
        localStorage.removeItem('authToken')
        localStorage.removeItem('user')
        return navigateTo('/auth/login')
    }
})
