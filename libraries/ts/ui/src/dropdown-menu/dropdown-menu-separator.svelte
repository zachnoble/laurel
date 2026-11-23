<script lang="ts">
	import { DropdownMenu } from 'bits-ui'

	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: DropdownMenu.SeparatorProps = $props()
</script>

<DropdownMenu.Separator
	bind:ref
	class={cn(
		'border-t-border/50 my-1 w-[calc(100%-(--spacing(4)))] self-center border-t',
		className,
	)}
	{...restProps}
/>
