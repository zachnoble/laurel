<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let {
		class: className,
		selected = false,
		children,
		...restProps
	}: HTMLAttributes<HTMLTableRowElement> & { selected?: boolean } = $props()
</script>

<tr
	aria-selected={selected}
	data-state={selected ? 'selected' : undefined}
	class={cn(
		'group/row relative text-sm select-none',
		'in-[tbody]:hover:bg-zinc-100 dark:in-[tbody]:hover:bg-zinc-700',
		'data-[state=selected]:bg-accent/5 dark:data-[state=selected]:bg-accent/35',
		className,
	)}
	{...restProps}
>
	{@render children?.()}
</tr>
