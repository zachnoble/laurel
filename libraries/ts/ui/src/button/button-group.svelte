<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let {
		class: className,
		orientation = 'horizontal',
		children,
		...restProps
	}: HTMLAttributes<HTMLDivElement> & { orientation?: 'horizontal' | 'vertical' } = $props()
</script>

<div
	data-ui="button-group"
	class={cn(
		'group inline-flex w-max items-center',
		'[&>*:not(:first-child):not(:last-child)]:rounded-none',
		orientation === 'horizontal'
			? '[&>*:first-child]:rounded-e-none [&>*:last-child]:rounded-s-none'
			: 'flex-col [&>*:first-child]:rounded-b-none [&>*:last-child]:rounded-t-none',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</div>
