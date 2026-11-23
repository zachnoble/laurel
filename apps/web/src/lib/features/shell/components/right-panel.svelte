<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import { Card } from '@laurel/ui/card/index.js'
	import { Input } from '@laurel/ui/input/index.js'
	import SearchIcon from '@lucide/svelte/icons/search'
	import { resolve } from '$app/paths'
	import { page } from '$app/state'

	import Logo from '#lib/features/shell/components/logo.svelte'
</script>

<aside class="sticky top-0 hidden h-dvh w-80 p-5.5 md:block lg:w-105">
	<div class="mb-4">
		<Input placeholder="Search..." icon={SearchIcon} />
	</div>
	{#if !page.data.user}
		<Card class="flex flex-col items-center">
			<Logo />
			<div class="mt-4 mb-5 text-center">
				<p class="text-xl font-bold">Login or sign up</p>
				<p class="text-muted mt-1 text-sm">
					Lorem ipsum dolor sit amet, consectetur adipiscing elit nam tincidunt lectus.
				</p>
			</div>
			<div class="flex w-full flex-col gap-3">
				<Button href={resolve('login')} class="w-full">Login</Button>
				<Button href="/explore" variant="outline" class="w-full">Explore more</Button>
			</div>
		</Card>
	{/if}
</aside>
