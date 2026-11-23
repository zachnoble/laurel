<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div class={cn('dark:bg-subtle overflow-hidden rounded-sm border p-5', className)} {...restProps}>
	{@render children?.()}
</div>
