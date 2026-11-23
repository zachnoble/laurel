<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import { page } from '$app/state'

	import type { NavItem } from '../nav-items'
	import { collapsibleLabel, navItem } from '../styles'

	let { item, collapsible = false }: { item: NavItem; collapsible?: boolean } = $props()

	const active = $derived(
		page.url.pathname === item.href ||
			(item.href !== '/' && page.url.pathname.startsWith(`${item.href}/`)),
	)
</script>

<Button
	href={item.href}
	variant="plain"
	aria-current={active ? 'page' : undefined}
	class={navItem({ active, collapsible })}
>
	<item.icon class="text-current" />
	<span class={collapsible ? collapsibleLabel : undefined}>{item.label}</span>
</Button>
