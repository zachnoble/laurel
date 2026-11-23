<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import { getCalloutContext } from './context'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()

	const { descriptionId } = getCalloutContext()
</script>

<div
	id={descriptionId}
	data-ui="callout-description"
	class={cn(
		'col-start-2 -col-end-2 text-sm/6 text-pretty text-(--callout-description)',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</div>
