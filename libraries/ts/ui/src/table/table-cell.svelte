<script lang="ts">
	import type { HTMLTdAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'

	let { class: className, children, ...restProps }: HTMLTdAttributes = $props()
</script>

<td class={cn('truncate border-b p-2 group-last/row:border-b-0', className)} {...restProps}>
	{@render children?.()}
</td>
