export const useApi = () => {
    const baseURL = 'http://localhost:8080/api'

    const apiFetch = async <T>(endpoint: string, options: any = {}): Promise<T> => {
        const token = localStorage.getItem('authToken')

        return await $fetch<T>(`${baseURL}${endpoint}`, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...(token ? { Authorization: `Bearer ${token}` } : {}),
                ...options.headers
            }
        })
    }

    return {
        baseURL,
        fetch: apiFetch
    }
}