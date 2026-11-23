<script lang="ts">
	import { cn } from '@laurel/ui/cn.js'
	import BoxIcon from '@lucide/svelte/icons/box'

	let { class: className }: { class?: string } = $props()
</script>

<BoxIcon class={cn('text-accent size-9 shrink-0', className)} aria-hidden="true" />
