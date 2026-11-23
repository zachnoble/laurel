<script lang="ts" module>
	import type { Button as ButtonPrimitive } from 'bits-ui'
	import { tv, type VariantProps } from 'tailwind-variants/lite'

	export const buttonVariants = tv({
		base: [
			'group relative inline-flex items-center justify-center gap-x-2 rounded-md font-semibold',
			'focus-visible:outline-ring outline-hidden focus-visible:outline-2 focus-visible:outline-offset-2',
			'disabled:opacity-50 aria-disabled:pointer-events-none aria-disabled:opacity-50',
			'text-sm/6 [&_svg:not([class*=size-])]:size-(--icon-size)',
			'[&_svg:not([class*=text-])]:text-(--icon-color)',
			'[--btn-bg:var(--color-accent)] [--btn-color:var(--color-foreground)]',
		],
		variants: {
			color: {
				foreground: '',
				accent: '[--btn-bg:var(--color-accent)] [--btn-color:var(--color-accent)]',
				red: '[--btn-bg:var(--color-red-600)] [--btn-color:var(--color-red-600)] dark:[--btn-color:var(--color-red-400)]',
				green: '[--btn-bg:var(--color-green-600)] [--btn-color:var(--color-green-600)] dark:[--btn-color:var(--color-green-400)]',
			},
			variant: {
				solid: [
					'bg-(--btn-bg) text-white [--btn-color:white] [--icon-color:white]',
					'ring-1 ring-white/15 ring-inset dark:ring-[oklch(from_var(--btn-bg)_calc(l*1.15)_c_h)]',
					'not-disabled:hover:opacity-90',
				],
				outline: [
					'shadow-outline text-(--btn-color) in-data-[ui=button-group]:shadow-none',
					'bg-white dark:bg-white/5',
					'not-disabled:hover:bg-subtle! dark:not-disabled:hover:bg-zinc-800!',
				],
				plain: 'not-disabled:hover:bg-subtle text-(--btn-color)',
				link: 'text-muted font-normal underline underline-offset-4 not-hover:decoration-(--btn-color)/15',
			},
			size: {
				sm: '',
				md: '[--icon-size:--spacing(4)]',
				lg: '[--icon-size:--spacing(5)]',
			},
			iconOnly: {
				true: '',
				false: '[--icon-color:var(--btn-color)]',
			},
			pending: {
				true: 'text-transparent',
				false: '',
			},
		},
		compoundVariants: [
			{ size: 'sm', iconOnly: true, class: 'size-7 [--icon-size:--spacing(4)]' },
			{
				size: 'sm',
				iconOnly: false,
				variant: ['solid', 'outline', 'plain'],
				class: 'h-7 px-2 text-xs/6 [--icon-size:--spacing(3)]',
			},
			{ size: 'md', iconOnly: true, class: 'p-1.5 [&_svg]:m-1' },
			{
				size: 'md',
				iconOnly: false,
				variant: ['solid', 'outline', 'plain'],
				class: 'px-3 py-1.5',
			},
			{ size: 'lg', iconOnly: true, class: 'p-2 [&_svg]:m-0.5' },
			{
				size: 'lg',
				iconOnly: false,
				variant: ['solid', 'outline', 'plain'],
				class: 'px-3.5 py-2',
			},
		],
		defaultVariants: {
			color: 'foreground',
			variant: 'solid',
			size: 'md',
			iconOnly: false,
			pending: false,
		},
	})

	export type ButtonVariants = VariantProps<typeof buttonVariants>

	export type ButtonProps = ButtonPrimitive.RootProps &
		Omit<ButtonVariants, 'pending'> & {
			pending?: boolean
		}
</script>

<script lang="ts">
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle'
	import { Button } from 'bits-ui'

	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		class: className,
		variant = 'solid',
		color = 'foreground',
		size = 'md',
		iconOnly = false,
		pending = false,
		disabled,
		children,
		...restProps
	}: ButtonProps = $props()
</script>

<Button.Root
	bind:ref
	data-variant={variant}
	class={cn(buttonVariants({ variant, color, size, iconOnly, pending }), className)}
	disabled={disabled || pending}
	aria-busy={pending || undefined}
	{...restProps}
>
	{#if pending}
		<LoaderCircleIcon class="absolute animate-spin" aria-label="Loading" />
		<span class="contents" aria-hidden="true">{@render children?.()}</span>
	{:else}
		{@render children?.()}
	{/if}
</Button.Root>
