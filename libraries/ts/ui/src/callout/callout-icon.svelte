<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div
	data-ui="callout-icon"
	class={cn(
		'col-start-1 me-2 flex items-start sm:row-start-1 sm:row-end-(--callout-content-row-end)',
		'[&:has(+[data-ui=callout-heading]+[data-ui=callout-description])]:pt-0.5',
		'[&:not(:has(+[data-ui=callout-heading]+[data-ui=callout-description]))]:self-center',
		'[&_svg]:size-5 [&_svg:not([class*=text-])]:text-(--callout-icon)',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</div>
