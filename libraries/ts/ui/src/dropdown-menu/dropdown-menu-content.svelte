<script lang="ts">
	import { DropdownMenu } from 'bits-ui'

	import { cn } from '../cn.js'
	import { floatingSurface } from '../popover/index.js'

	let {
		ref = $bindable(null),
		class: className,
		sideOffset = 8,
		...restProps
	}: DropdownMenu.ContentProps = $props()
</script>

<DropdownMenu.Portal>
	<DropdownMenu.Content
		bind:ref
		{sideOffset}
		class={cn(
			floatingSurface,
			'flex max-h-(--bits-dropdown-menu-content-available-height) max-w-72 min-w-[max(--spacing(48),var(--bits-dropdown-menu-anchor-width))] flex-col overflow-auto p-1',
			className,
		)}
		{...restProps}
	/>
</DropdownMenu.Portal>
