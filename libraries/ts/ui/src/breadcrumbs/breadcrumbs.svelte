<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLAttributes<HTMLElement> = $props()
</script>

<nav aria-label="Breadcrumbs" {...restProps}>
	<ol class={cn('flex gap-1', className)}>
		{@render children?.()}
	</ol>
</nav>
