<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left'
	import { resolve } from '$app/paths'

	import type { LayoutProps } from './$types'

	let { children }: LayoutProps = $props()
</script>

<main class="flex min-h-dvh w-full items-center justify-center py-0 text-lg sm:py-16">
	<div class="mx-auto w-screen sm:max-w-118.75">
		<div class="absolute top-3 left-3 sm:top-5 sm:left-5">
			<Button href={resolve('')} variant="plain" class="pr-3">
				<ChevronLeftIcon /> Home
			</Button>
		</div>
		<div
			class="bg-background sm:dark:bg-subtle rounded-none border-0 p-8 sm:rounded-sm sm:border sm:p-12"
		>
			{@render children()}
		</div>
	</div>
</main>
