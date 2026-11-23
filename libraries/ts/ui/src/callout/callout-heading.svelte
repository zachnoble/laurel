<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import { getCalloutContext } from './context'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()

	const { headingId } = getCalloutContext()
</script>

<div
	id={headingId}
	data-ui="callout-heading"
	class={cn(
		'col-start-2 -col-end-2 flex gap-x-2 gap-y-1 self-center text-sm/6 font-medium text-(--callout-heading)',
		'sm:[&+[data-ui=callout-description]]:leading-5 [&:has(+[data-ui=callout-description])]:pb-0.5',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</div>
