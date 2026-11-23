<script lang="ts" module>
	import { tv, type VariantProps } from 'tailwind-variants/lite'

	export const calloutVariants = tv({
		base: [
			'group grid w-full grid-cols-[auto_1fr_auto_auto] rounded-lg border p-4',
			'border-(--callout-border) bg-(--callout-bg)',
			'[--callout-border:var(--color-border)] [--callout-icon:var(--color-muted)]',
			'[&:has([data-ui=callout-heading]+[data-ui=callout-description])]:[--callout-content-row-end:3]',
		],
		variants: {
			color: {
				default:
					'shadow-xs [&:has(>[data-ui=callout-heading])]:[--callout-description:var(--color-muted)]',
				zinc: [
					'[--callout-bg:var(--color-subtle)] [--callout-description:var(--color-foreground)]/80',
					'[--callout-icon:var(--color-zinc-600)] dark:[--callout-icon:var(--color-zinc-400)]',
				],
				red: [
					'[--callout-bg:var(--color-red-50)] [--callout-border:var(--color-red-200)]',
					'[--callout-description:var(--color-red-700)] [--callout-heading:var(--color-red-700)] [--callout-icon:var(--color-red-600)]',
					'dark:[--callout-bg:var(--color-red-400)]/10 dark:[--callout-border:var(--color-red-400)]/50',
					'dark:[--callout-description:var(--color-red-300)] dark:[--callout-heading:var(--color-red-200)] dark:[--callout-icon:var(--color-red-400)]',
				],
				blue: [
					'[--callout-bg:var(--color-blue-50)] [--callout-border:var(--color-blue-200)]',
					'[--callout-description:var(--color-blue-600)] [--callout-heading:var(--color-blue-600)] [--callout-icon:var(--color-blue-500)]',
					'dark:[--callout-bg:var(--color-blue-400)]/10 dark:[--callout-border:var(--color-blue-400)]/50',
					'dark:[--callout-description:var(--color-blue-300)] dark:[--callout-heading:var(--color-blue-200)] dark:[--callout-icon:var(--color-blue-400)]',
				],
				yellow: [
					'[--callout-bg:var(--color-yellow-50)] [--callout-border:var(--color-yellow-300)]',
					'[--callout-description:var(--color-yellow-700)] [--callout-heading:var(--color-yellow-700)] [--callout-icon:var(--color-yellow-600)]',
					'dark:[--callout-bg:var(--color-yellow-400)]/10 dark:[--callout-border:var(--color-yellow-400)]/50',
					'dark:[--callout-description:var(--color-yellow-300)] dark:[--callout-heading:var(--color-yellow-200)] dark:[--callout-icon:var(--color-yellow-400)]',
				],
				green: [
					'[--callout-bg:var(--color-green-50)] [--callout-border:var(--color-green-200)]',
					'[--callout-description:var(--color-green-700)] [--callout-heading:var(--color-green-700)] [--callout-icon:var(--color-green-600)]',
					'dark:[--callout-bg:var(--color-green-400)]/10 dark:[--callout-border:var(--color-green-400)]/50',
					'dark:[--callout-description:var(--color-green-300)] dark:[--callout-heading:var(--color-green-200)] dark:[--callout-icon:var(--color-green-400)]',
				],
			},
		},
		defaultVariants: { color: 'default' },
	})

	export type CalloutColor = NonNullable<VariantProps<typeof calloutVariants>['color']>
</script>

<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import { setCalloutContext } from './context'

	let {
		class: className,
		color = 'default',
		children,
		...restProps
	}: Omit<HTMLAttributes<HTMLDivElement>, 'color'> & { color?: CalloutColor } = $props()

	const id = $props.id()
	const headingId = `${id}-heading`
	const descriptionId = `${id}-description`

	setCalloutContext({ headingId, descriptionId })
</script>

<div
	role="note"
	aria-labelledby={restProps['aria-label'] ? undefined : headingId}
	aria-describedby={descriptionId}
	class={cn(calloutVariants({ color }), className)}
	{...restProps}
>
	{@render children?.()}
</div>
