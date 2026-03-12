import { ref } from 'vue'

export type ToastType = 'error' | 'success' | 'info'

interface Toast {
    id: number
    type: ToastType
    message: string
}

// Module-level singleton so all components share the same toast state
const toasts = ref<Toast[]>([])
let nextId = 0

export function useToast() {
    function dismiss(id: number) {
        toasts.value = toasts.value.filter(t => t.id !== id)
    }

    function add(type: ToastType, message: string, duration = 4000) {
        const id = nextId++
        toasts.value.push({ id, type, message })
        setTimeout(() => dismiss(id), duration)
    }

    return {
        toasts,
        dismiss,
        error: (message: string) => add('error', message),
        success: (message: string) => add('success', message),
        info: (message: string) => add('info', message),
    }
}
