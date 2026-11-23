<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import MenuIcon from '@lucide/svelte/icons/menu'
	import XIcon from '@lucide/svelte/icons/x'
	import { createMutation } from '@tanstack/svelte-query'
	import { Dialog } from 'bits-ui'
	import { MediaQuery } from 'svelte/reactivity'
	import { afterNavigate, goto } from '$app/navigation'
	import { resolve } from '$app/paths'
	import { page } from '$app/state'

	import Logo from '#lib/features/shell/components/logo.svelte'
	import { authClient } from '#lib/rpc-clients.js'
	import { navItems } from '../nav-items'
	import NavLink from './nav-link.svelte'

	let open = $state(false)

	afterNavigate(() => {
		open = false
	})

	const desktop = new MediaQuery('min-width: 48rem')
	$effect(() => {
		if (desktop.current) open = false
	})

	const signOutMutation = createMutation(() => ({
		mutationFn: () => authClient.signOut({}),
		onSuccess: () => goto(resolve('login'), { invalidateAll: true }),
	}))
</script>

{#snippet header(action: 'open' | 'close')}
	<div class="flex items-center justify-between px-4 py-3">
		<Logo />
		{#if action === 'open'}
			<Dialog.Trigger class="rounded-lg p-1" aria-label="Open navigation">
				<MenuIcon class="text-muted size-6" />
			</Dialog.Trigger>
		{:else}
			<Dialog.Close class="rounded-lg p-1" aria-label="Close navigation">
				<XIcon class="text-muted size-6" />
			</Dialog.Close>
		{/if}
	</div>
{/snippet}

<Dialog.Root bind:open>
	<div class="bg-background fixed inset-x-0 top-0 z-30 md:hidden">
		{@render header('open')}
	</div>

	<Dialog.Portal>
		<Dialog.Content class="bg-background fixed inset-0 z-40 flex flex-col md:hidden">
			<Dialog.Title class="sr-only">Navigation</Dialog.Title>
			{@render header('close')}
			<nav class="mt-6 flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto px-4">
				{#each navItems as item (item.href)}
					<NavLink {item} />
				{/each}
			</nav>
			<div class="flex shrink-0 flex-col gap-4 px-4 pb-5.5">
				<Button class="w-full">Create</Button>
				{#if page.data.user}
					<Button
						variant="outline"
						class="w-full"
						pending={signOutMutation.isPending}
						onclick={() => signOutMutation.mutate()}
					>
						Logout
					</Button>
				{:else}
					<Button href={resolve('login')} variant="outline" class="w-full">Login</Button>
				{/if}
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
