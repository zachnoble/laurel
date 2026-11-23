<script lang="ts">
	import { DropdownMenu } from 'bits-ui'

	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		class: className,
		destructive = false,
		...restProps
	}: DropdownMenu.ItemProps & { destructive?: boolean } = $props()
</script>

<DropdownMenu.Item
	bind:ref
	class={cn(
		'group flex cursor-pointer items-center gap-x-2 rounded-sm px-2 py-2.5 text-sm/6 outline-hidden select-none sm:py-1.5',
		'data-highlighted:bg-subtle data-disabled:opacity-50',
		'[&:not([data-highlighted])_svg:not([class*=text-])]:text-muted/75 [&_svg:not([class*=size-])]:size-4',
		destructive && 'text-red-600 [&_svg]:text-red-600!',
		className,
	)}
	{...restProps}
/>
