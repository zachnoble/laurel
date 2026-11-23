<script lang="ts">
	import { Checkbox } from 'bits-ui'

	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		value = $bindable([]),
		class: className,
		children,
		...restProps
	}: Checkbox.GroupProps = $props()
</script>

<Checkbox.Group bind:ref bind:value class={cn('flex flex-col', className)} {...restProps}>
	<div data-ui="box" class="mt-3 flex flex-col gap-y-3">
		{@render children?.()}
	</div>
</Checkbox.Group>
