import { useRouter } from 'vue-router'
import { useAuth } from '~/composables/useAuth'

export const useApi = () => {
    const baseURL = 'http://localhost:8080/api'
    const router = useRouter()
    const auth = useAuth()

    const apiFetch = async <T>(endpoint: string, options: any = {}): Promise<T> => {
        const token = localStorage.getItem('authToken')

        try {
            return await $fetch<T>(`${baseURL}${endpoint}`, {
                ...options,
                headers: {
                    'Content-Type': 'application/json',
                    ...(token ? { Authorization: `Bearer ${token}` } : {}),
                    ...options.headers
                }
            })
        } catch (err: any) {
            if (err?.status === 401 || err?.statusCode === 401) {
                auth.clearUser()
                router.push('/auth/login')
            }
            throw err
        }
    }

    return {
        baseURL,
        fetch: apiFetch
    }
}
