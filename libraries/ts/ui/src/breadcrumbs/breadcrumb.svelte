<script lang="ts">
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right'
	import type { HTMLAnchorAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let {
		class: className,
		href,
		current = false,
		children,
		...restProps
	}: HTMLAnchorAttributes & { current?: boolean } = $props()
</script>

<li class={cn('group/crumb flex items-center gap-1', className)}>
	<a
		{href}
		class="underline underline-offset-4"
		aria-current={current ? 'page' : undefined}
		{...restProps}
	>
		{@render children?.()}
	</a>
	<ChevronRightIcon class="text-muted size-4 group-last/crumb:hidden" aria-hidden="true" />
</li>
