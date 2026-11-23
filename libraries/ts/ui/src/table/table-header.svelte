<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let {
		class: className,
		children,
		...restProps
	}: HTMLAttributes<HTMLTableSectionElement> = $props()
</script>

<thead class={cn('sticky top-0 z-10 backdrop-blur-md', className)} {...restProps}>
	{@render children?.()}
</thead>
