<script lang="ts" module>
	export type ToastData = {
		type?: 'success' | 'error' | 'info' | 'warning'
		title?: string
		message: string
	}
</script>

<script lang="ts">
	import CircleCheckIcon from '@lucide/svelte/icons/circle-check'
	import CircleXIcon from '@lucide/svelte/icons/circle-x'
	import InfoIcon from '@lucide/svelte/icons/info'
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert'
	import XIcon from '@lucide/svelte/icons/x'

	let { data, closeToast }: { data: ToastData; closeToast?: () => void } = $props()

	const config = {
		success: { icon: CircleCheckIcon, title: 'Success', class: 'text-green-400' },
		error: { icon: CircleXIcon, title: 'Error', class: 'text-red-400' },
		info: { icon: InfoIcon, title: 'Info', class: 'text-rose-400' },
		warning: { icon: TriangleAlertIcon, title: 'Warning', class: 'text-yellow-500' },
	}

	const type = $derived(config[data.type ?? 'info'])
</script>

<div
	class="fade-in animate-in border-border/80 bg-background pointer-events-auto rounded-md border duration-300"
>
	<div class="flex h-full items-center">
		<div class="flex items-center px-3 py-2.5">
			<type.icon class="size-6 fill-current stroke-(--color-background) {type.class}" />
			<div class="mx-3 max-w-68.75 min-w-52.5">
				<p class="pb-0.5 text-sm font-medium">{data.title ?? type.title}</p>
				<p class="text-muted text-xs">{data.message}</p>
			</div>
		</div>

		<button
			type="button"
			aria-label="Dismiss"
			onclick={() => closeToast?.()}
			class="text-muted hover:bg-muted/3 mr-2 rounded-full p-2.5 text-xs font-medium"
		>
			<XIcon class="size-3.5" />
		</button>
	</div>
</div>
