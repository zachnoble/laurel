<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, ...restProps }: HTMLAttributes<HTMLDivElement> = $props()
</script>

<div
	class={cn('animate-pulse rounded-md bg-zinc-200 dark:bg-zinc-700', className)}
	{...restProps}
></div>
