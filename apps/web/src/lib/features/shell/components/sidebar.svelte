<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import PlusIcon from '@lucide/svelte/icons/plus'
	import { page } from '$app/state'

	import Logo from '#lib/features/shell/components/logo.svelte'
	import { navItems } from '../nav-items'
	import { collapsibleLabel } from '../styles'
	import MoreMenu from './more-menu.svelte'
	import NavLink from './nav-link.svelte'
</script>

<div class="sticky top-0 hidden h-dvh w-20 flex-col justify-between p-4.5 md:flex lg:w-60 lg:p-5.5">
	<div class="flex flex-col items-center lg:items-stretch">
		<Logo />
		<nav class="flex w-full flex-col gap-2 py-6.5">
			{#each navItems as item (item.href)}
				<NavLink {item} collapsible />
			{/each}
		</nav>
		<Button class="max-lg:p-2 lg:w-full">
			<PlusIcon class="size-5 lg:hidden" />
			<span class={collapsibleLabel}>Create</span>
		</Button>
	</div>
	{#if page.data.user}
		<MoreMenu />
	{/if}
</div>
