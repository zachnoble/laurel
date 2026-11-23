<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import {
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuRoot,
		DropdownMenuSeparator,
		DropdownMenuTrigger,
	} from '@laurel/ui/dropdown-menu/index.js'
	import ArrowRightIcon from '@lucide/svelte/icons/arrow-right'
	import CircleHelpIcon from '@lucide/svelte/icons/circle-help'
	import MenuIcon from '@lucide/svelte/icons/menu'
	import SettingsIcon from '@lucide/svelte/icons/settings'
	import { createMutation } from '@tanstack/svelte-query'
	import { goto } from '$app/navigation'
	import { resolve } from '$app/paths'

	import { authClient } from '#lib/rpc-clients.js'
	import { collapsibleLabel, navItem } from '../styles'

	const signOutMutation = createMutation(() => ({
		mutationFn: () => authClient.signOut({}),
		onSuccess: () => goto(resolve('login'), { invalidateAll: true }),
	}))

	const links = [
		{ label: 'Settings', href: '/settings', icon: SettingsIcon },
		{ label: 'Support', href: '/support', icon: CircleHelpIcon },
	]
</script>

<DropdownMenuRoot>
	<DropdownMenuTrigger>
		{#snippet child({ props })}
			<Button {...props} variant="plain" class={navItem({ collapsible: true })}>
				<MenuIcon class="text-current" />
				<span class={collapsibleLabel}>More</span>
			</Button>
		{/snippet}
	</DropdownMenuTrigger>
	<DropdownMenuContent side="top" align="start">
		{#each links as link (link.href)}
			<DropdownMenuItem onSelect={() => goto(link.href)}>
				<link.icon />
				{link.label}
			</DropdownMenuItem>
		{/each}
		<DropdownMenuSeparator />
		<DropdownMenuItem
			onSelect={() => signOutMutation.mutate()}
			disabled={signOutMutation.isPending}
		>
			<ArrowRightIcon />
			Sign out
		</DropdownMenuItem>
	</DropdownMenuContent>
</DropdownMenuRoot>
