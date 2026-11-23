<script lang="ts">
	import type { Snippet } from 'svelte'

	import { cn } from '../cn.js'

	let {
		heading,
		description,
		class: className,
		children,
	}: { heading: string; description?: string; class?: string; children?: Snippet } = $props()
</script>

<div
	class={cn(
		'@container flex h-full w-full flex-col items-center justify-center gap-1 p-4 text-center',
		className,
	)}
>
	<div class="text-base font-semibold text-balance">{heading}</div>
	{#if description}
		<p class="text-muted max-w-prose text-sm/6 text-balance">{description}</p>
	{/if}
	{@render children?.()}
</div>
