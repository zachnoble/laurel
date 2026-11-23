import { tv } from 'tailwind-variants/lite'

export const navItem = tv({
	base: 'flex w-full items-center gap-3.5 rounded-lg p-3 font-medium [&_svg]:size-5 [&_svg]:shrink-0',
	variants: {
		active: {
			true: 'bg-subtle text-foreground',
			false: 'text-muted',
		},
		collapsible: {
			true: 'justify-center lg:justify-start',
			false: 'justify-start',
		},
	},
	defaultVariants: { active: false, collapsible: false },
})

export const collapsibleLabel = 'sr-only lg:not-sr-only'
