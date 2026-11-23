<script lang="ts">
	import NProgress from 'nprogress'
	import { navigating, page } from '$app/state'

	NProgress.configure({
		minimum: 0.2,
		easing: 'linear',
		speed: 200,
		trickle: true,
		trickleSpeed: 200,
		showSpinner: false,
	})

	const active = $derived(
		navigating.to !== null && navigating.to.url.pathname !== page.url.pathname,
	)

	$effect(() => {
		if (active) NProgress.start()
		else NProgress.done()
	})
</script>

<style>
	:global(#nprogress) {
		pointer-events: none;
	}

	:global(#nprogress .bar) {
		background: var(--color-accent);
		position: fixed;
		z-index: 1031;
		top: 0;
		left: 0;
		width: 100%;
		height: 2px;
	}
</style>
