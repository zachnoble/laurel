<script lang="ts">
	import { Tabs } from 'bits-ui'

	import { cn } from '../cn.js'
	import { getTabsContext } from './context'

	let { ref = $bindable(null), class: className, ...restProps }: Tabs.TriggerProps = $props()

	const tabs = getTabsContext()

	const styles = {
		underline: {
			horizontal:
				'before:bg-accent p-2 before:absolute before:inset-x-0 before:bottom-[-1.5px] data-[state=active]:before:h-0.5',
			vertical:
				'before:bg-accent px-4 before:absolute before:inset-y-0 before:left-[-1.5px] data-[state=active]:before:w-0.5',
		},
		pills: {
			horizontal:
				'px-3 py-2 data-[state=active]:bg-zinc-100 dark:data-[state=active]:bg-zinc-600/45',
			vertical:
				'px-3 py-2 data-[state=active]:bg-zinc-100 dark:data-[state=active]:bg-zinc-600/45',
		},
		segment: {
			horizontal:
				'flex-1 justify-center px-6 py-1 data-[state=active]:bg-background data-[state=active]:shadow-2xs dark:data-[state=active]:bg-zinc-500 [&_svg:not([class*=size-])]:size-4',
			vertical: '',
		},
	}
</script>

<Tabs.Trigger
	bind:ref
	class={cn(
		'text-muted relative flex cursor-pointer items-center gap-x-3 rounded-md font-medium outline-hidden',
		'hover:text-foreground data-[state=active]:text-foreground data-disabled:opacity-50',
		'focus-visible:ring-ring focus-visible:ring-2 focus-visible:ring-inset',
		'[&_svg:not([class*=size-])]:size-5',
		styles[tabs.variant][tabs.orientation],
		className,
	)}
	{...restProps}
/>
