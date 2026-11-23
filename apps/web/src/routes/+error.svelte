<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import { resolve } from '$app/paths'
	import { page } from '$app/state'

	const notFound = $derived(page.status === 404)
	const title = $derived(notFound ? '404 Not Found' : 'Something went wrong...')
	const description = $derived(
		notFound
			? 'The page you are looking for no longer exists.'
			: 'An unexpected error occurred. Please try refreshing the page.',
	)
</script>

<svelte:head>
	<title>{title}</title>
	<meta name="description" content={description}>
</svelte:head>

<main class="flex min-h-dvh w-screen flex-col items-center justify-center gap-6 py-12">
	<div class="flex flex-col items-center gap-4 text-center">
		<h1 class="text-3xl font-bold md:text-6xl">{title}</h1>
		<p class="text-muted max-w-md text-sm">{description}</p>
	</div>
	<Button href={resolve('')}>Return Home</Button>
</main>
