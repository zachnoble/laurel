import { toastError } from '@laurel/ui/toast/index.js'
import { MutationCache, QueryClient } from '@tanstack/svelte-query'

export const queryClient = new QueryClient({
	mutationCache: new MutationCache({
		onError: toastError,
	}),
})
