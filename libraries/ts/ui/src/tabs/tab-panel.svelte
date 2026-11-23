<script lang="ts">
	import { Tabs } from 'bits-ui'

	import { cn } from '../cn.js'
	import { getTabsContext } from './context'

	let { ref = $bindable(null), class: className, ...restProps }: Tabs.ContentProps = $props()

	const tabs = getTabsContext()

	const styles = {
		underline: { horizontal: 'py-4', vertical: 'px-4' },
		pills: { horizontal: 'px-5 py-4', vertical: 'p-4 sm:py-0 sm:pl-8' },
		segment: { horizontal: 'px-3 py-4', vertical: '' },
	}
</script>

<Tabs.Content
	bind:ref
	class={cn('flex-1 outline-hidden', styles[tabs.variant][tabs.orientation], className)}
	{...restProps}
/>
