<script lang="ts">
	import { Tabs } from 'bits-ui'

	import { cn } from '../cn.js'
	import { setTabsContext, type TabsVariant } from './context'

	let {
		ref = $bindable(null),
		value = $bindable(''),
		variant = 'underline',
		orientation = 'horizontal',
		activationMode = 'manual',
		class: className,
		...restProps
	}: Tabs.RootProps & { variant?: TabsVariant } = $props()

	setTabsContext({
		get variant() {
			return variant
		},
		get orientation() {
			return orientation
		},
	})
</script>

<Tabs.Root
	bind:ref
	bind:value
	{orientation}
	{activationMode}
	class={cn(
		'group flex',
		orientation === 'horizontal' ? 'flex-col' : 'flex-col sm:flex-row',
		className,
	)}
	{...restProps}
/>
