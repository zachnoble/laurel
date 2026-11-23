import { ConnectError } from '@connectrpc/connect'
import { toast } from 'svelte-sonner'

import Toast, { type ToastData } from './toast.svelte'

export { Toast, type ToastData }

export const sendToast = (data: ToastData) => {
	toast.custom(Toast, { componentProps: { data } })
}

export const toastError = (error: unknown) => {
	const message =
		error instanceof ConnectError
			? error.rawMessage
			: error instanceof Error
				? error.message
				: 'An unexpected error occurred'
	sendToast({ type: 'error', message })
}
