<script lang="ts">
	import MobileNav from '#lib/features/shell/components/mobile-nav.svelte'
	import RightPanel from '#lib/features/shell/components/right-panel.svelte'
	import Sidebar from '#lib/features/shell/components/sidebar.svelte'
	import type { LayoutProps } from './$types'

	let { children }: LayoutProps = $props()
</script>

<div class="flex w-full justify-between">
	<Sidebar />
	<MobileNav />
	<main class="mt-14 min-w-0 flex-1 md:mt-0">
		{@render children()}
	</main>
	<RightPanel />
</div>
