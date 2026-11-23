<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div
	data-ui="dialog-footer"
	class={cn('mt-auto flex flex-col-reverse justify-end gap-3 p-6 sm:flex-row', className)}
	{...restProps}
>
	{@render children?.()}
</div>
