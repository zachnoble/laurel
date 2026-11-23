<script lang="ts">
	import type { HTMLThAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLThAttributes = $props()
</script>

<th
	class={cn(
		'relative border-b p-2 text-start text-sm font-semibold',
		'not-last:after:bg-border not-last:after:absolute not-last:after:top-2 not-last:after:right-0 not-last:after:h-5 not-last:after:w-[1.5px] not-last:after:rounded-sm',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</th>
