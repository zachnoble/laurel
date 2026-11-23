<script lang="ts">
	import { Tabs } from 'bits-ui'

	import { cn } from '../cn.js'
	import { getTabsContext } from './context'

	let { ref = $bindable(null), class: className, ...restProps }: Tabs.ListProps = $props()

	const tabs = getTabsContext()

	const styles = {
		underline: {
			horizontal: 'w-full space-x-4 border-b',
			vertical: 'space-y-3.5 self-start border-l',
		},
		pills: { horizontal: 'space-x-4', vertical: 'space-y-2' },
		segment: { horizontal: 'bg-subtle rounded-lg p-0.5 shadow-xs', vertical: '' },
	}
</script>

<div class="flex overflow-x-auto pb-px pl-px">
	<Tabs.List
		bind:ref
		class={cn(
			'flex text-sm/6',
			tabs.orientation === 'horizontal'
				? 'whitespace-nowrap'
				: 'flex-1 flex-col sm:flex-initial',
			styles[tabs.variant][tabs.orientation],
			className,
		)}
		{...restProps}
	/>
</div>
