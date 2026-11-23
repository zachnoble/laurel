<script lang="ts">
	import type { HTMLTableAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import type { WithElementRef } from '../types.js'

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLTableAttributes, HTMLTableElement> = $props()
</script>

<div class="relative max-h-70 w-full scroll-pt-9 overflow-auto rounded-md border">
	<table
		bind:this={ref}
		class={cn('w-full border-separate border-spacing-0', className)}
		{...restProps}
	>
		{@render children?.()}
	</table>
</div>
