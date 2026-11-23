<script lang="ts">
	import { AlertDialog } from 'bits-ui'

	import { cn } from '../cn.js'
	import { modalContent, modalOverlay, modalPositioner } from '../dialog/styles.js'

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: AlertDialog.ContentProps = $props()
</script>

<AlertDialog.Portal>
	<AlertDialog.Overlay class={cn(modalOverlay)} />
	<div class={cn(modalPositioner, 'grid-rows-[1fr_auto_1fr] p-4 sm:grid-rows-[1fr_auto_3fr]')}>
		<AlertDialog.Content
			bind:ref
			class={cn(
				modalContent,
				'max-h-[calc(100dvh-32px)] rounded-xl sm:max-w-md',
				'data-[state=open]:zoom-in-95 data-[state=closed]:zoom-out-95',
				className,
			)}
			{...restProps}
		/>
	</div>
</AlertDialog.Portal>
