<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x'
	import { Dialog, type WithoutChildrenOrChild } from 'bits-ui'
	import type { Snippet } from 'svelte'

	import { Button } from '../button/index.js'
	import { cn } from '../cn.js'
	import { modalContent, modalOverlay, modalPositioner } from './styles'

	let {
		ref = $bindable(null),
		class: className,
		showCloseButton = true,
		children,
		...restProps
	}: WithoutChildrenOrChild<Dialog.ContentProps> & {
		showCloseButton?: boolean
		children?: Snippet
	} = $props()
</script>

<Dialog.Portal>
	<Dialog.Overlay class={cn(modalOverlay)} />
	<div
		class={cn(modalPositioner, 'grid-rows-[1fr_auto] pt-4 sm:grid-rows-[1fr_auto_3fr] sm:p-4')}
	>
		<Dialog.Content
			bind:ref
			class={cn(
				modalContent,
				'max-h-[calc(100dvh-16px)] rounded-t-xl sm:max-h-[calc(100dvh-32px)] sm:max-w-lg sm:rounded-xl',
				'data-[state=open]:slide-in-from-bottom sm:data-[state=open]:slide-in-from-bottom-0 sm:data-[state=open]:zoom-in-95',
				'data-[state=closed]:slide-out-to-bottom sm:data-[state=closed]:slide-out-to-bottom-0 sm:data-[state=closed]:zoom-out-95',
				className,
			)}
			{...restProps}
		>
			{@render children?.()}
			{#if showCloseButton}
				<Dialog.Close>
					{#snippet child({ props })}
						<Button
							{...props}
							variant="plain"
							size="lg"
							iconOnly
							class="text-muted/75 hover:text-foreground absolute end-2 top-3 p-1.5"
						>
							<XIcon />
							<span class="sr-only">Close</span>
						</Button>
					{/snippet}
				</Dialog.Close>
			{/if}
		</Dialog.Content>
	</div>
</Dialog.Portal>
