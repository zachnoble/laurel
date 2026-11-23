<script lang="ts">
	import { Dialog } from 'bits-ui'

	import { cn } from '../cn.js'

	let { ref = $bindable(null), class: className, ...restProps }: Dialog.TitleProps = $props()
</script>

<Dialog.Title
	bind:ref
	data-ui="dialog-header"
	class={cn('ps-6 pe-10 pt-6 pb-2 text-base font-semibold', className)}
	{...restProps}
/>
