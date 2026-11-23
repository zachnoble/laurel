<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check'
	import MinusIcon from '@lucide/svelte/icons/minus'
	import { Checkbox, type WithoutChildrenOrChild } from 'bits-ui'
	import type { Snippet } from 'svelte'

	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		checked = $bindable(false),
		indeterminate = $bindable(false),
		class: className,
		children: label,
		...restProps
	}: WithoutChildrenOrChild<Checkbox.RootProps> & { children?: Snippet } = $props()

	const iconClass = 'size-4 text-white'
</script>

{#snippet box()}
	<Checkbox.Root
		bind:ref
		bind:checked
		bind:indeterminate
		class={cn(
			'flex size-4 shrink-0 cursor-pointer items-center justify-center rounded-sm border shadow-xs outline-hidden',
			'border-[oklch(from_var(--color-input)_calc(l*var(--contract,0.9))_c_h)] dark:bg-white/5 dark:[--contract:1.05]',
			'data-[state=checked]:border-accent data-[state=checked]:bg-accent',
			'data-[state=indeterminate]:border-accent data-[state=indeterminate]:bg-accent',
			'focus-visible:outline-ring focus-visible:outline-2 focus-visible:outline-offset-2',
			'aria-invalid:border-red-600 data-disabled:cursor-not-allowed',
			label && 'me-3',
			className,
		)}
		{...restProps}
	>
		{#snippet children({ checked, indeterminate })}
			{#if indeterminate}
				<MinusIcon class={iconClass} />
			{:else if checked}
				<CheckIcon class={iconClass} />
			{/if}
		{/snippet}
	</Checkbox.Root>
{/snippet}

{#if label}
	<!-- biome-ignore lint/a11y/noLabelWithoutControl: Checkbox.Root renders the control (button[role=checkbox]) -->
	<label class="flex cursor-pointer items-center text-sm/6 has-data-disabled:opacity-50">
		{@render box()}
		{@render label()}
	</label>
{:else}
	{@render box()}
{/if}
