<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div
	data-ui="dialog-body"
	class={cn('flex flex-1 flex-col overflow-auto px-6', className)}
	{...restProps}
>
	{@render children?.()}
</div>
