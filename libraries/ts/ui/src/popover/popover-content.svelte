<script lang="ts">
	import { Popover } from 'bits-ui'

	import { cn } from '../cn.js'
	import { floatingSurface } from './styles'

	let {
		ref = $bindable(null),
		class: className,
		sideOffset = 8,
		...restProps
	}: Popover.ContentProps = $props()
</script>

<Popover.Portal>
	<Popover.Content bind:ref {sideOffset} class={cn(floatingSurface, className)} {...restProps} />
</Popover.Portal>
