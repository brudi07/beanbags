import { ref } from 'vue'

interface ConfirmOptions {
    title?: string
    confirmLabel?: string
    cancelLabel?: string
    danger?: boolean
}

interface ConfirmState {
    message: string
    title: string
    confirmLabel: string
    cancelLabel: string
    danger: boolean
    resolve: (value: boolean) => void
}

const state = ref<ConfirmState | null>(null)

export function useConfirm() {
    function confirm(message: string, options: ConfirmOptions = {}): Promise<boolean> {
        return new Promise((resolve) => {
            state.value = {
                message,
                title: options.title ?? 'Are you sure?',
                confirmLabel: options.confirmLabel ?? 'Confirm',
                cancelLabel: options.cancelLabel ?? 'Cancel',
                danger: options.danger ?? true,
                resolve,
            }
        })
    }

    function respond(value: boolean) {
        state.value?.resolve(value)
        state.value = null
    }

    return { state, confirm, respond }
}
