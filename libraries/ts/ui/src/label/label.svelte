<script lang="ts">
	import { Label } from 'bits-ui'

	import { cn } from '../cn.js'

	let { ref = $bindable(null), class: className, ...restProps }: Label.RootProps = $props()
</script>

<Label.Root
	bind:ref
	data-ui="label"
	class={cn('inline-block min-w-max text-sm/6 font-medium text-pretty', className)}
	{...restProps}
/>
