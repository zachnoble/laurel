<script lang="ts">
	import type { Component } from 'svelte'
	import type { HTMLInputAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import type { WithElementRef } from '../types.js'

	type Props = WithElementRef<Omit<HTMLInputAttributes, 'type'>, HTMLInputElement> & {
		type?: Exclude<HTMLInputAttributes['type'], 'file' | 'checkbox' | 'radio'>
		label?: string
		errorMessage?: string
		icon?: Component<{ class?: string }>
	}

	const uid = $props.id()

	let {
		ref = $bindable(null),
		value = $bindable(),
		class: className,
		label,
		errorMessage,
		icon: Icon,
		id = uid,
		...restProps
	}: Props = $props()

	const errorId = $derived(`${id}-error`)
</script>

<div class="relative">
	{#if Icon}
		<Icon
			class="text-muted pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2"
		/>
	{/if}
	<input
		bind:this={ref}
		bind:value
		{id}
		aria-label={label ?? restProps.placeholder}
		aria-invalid={errorMessage ? true : undefined}
		aria-describedby={errorMessage ? errorId : undefined}
		class={cn(
			'border-input w-full rounded-md border px-3 py-[calc(--spacing(1.5)-1px)] shadow-xs outline-hidden',
			'placeholder:text-muted text-sm/6',
			'read-only:bg-zinc-800/5 dark:shadow-none dark:read-only:bg-white/10',
			'disabled:opacity-50 aria-invalid:border-red-600',
			'focus:border-ring focus:ring-ring focus:ring-1',
			Icon && 'pl-8',
			className,
		)}
		{...restProps}
	>
</div>
{#if errorMessage}
	<p id={errorId} class="mt-2 block text-sm/6 text-red-600">{errorMessage}</p>
{/if}
