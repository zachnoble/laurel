<script lang="ts">
	import './layout.css'

	import { QueryClientProvider } from '@tanstack/svelte-query'
	import { ModeWatcher, mode } from 'mode-watcher'
	import { onMount } from 'svelte'
	import { Toaster } from 'svelte-sonner'

	import NavProgress from '#lib/features/shell/components/nav-progress.svelte'
	import { queryClient } from '#lib/query-client.js'
	import type { LayoutProps } from './$types'

	let { children }: LayoutProps = $props()

	onMount(() => document.getElementById('app-skeleton')?.remove())
</script>

<ModeWatcher disableHeadScriptInjection />
<NavProgress />
<Toaster position="top-center" theme={mode.current} />

<QueryClientProvider client={queryClient}> {@render children()} </QueryClientProvider>
