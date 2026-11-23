<script lang="ts">
	import { DateFormatter, type DateValue, getLocalTimeZone } from '@internationalized/date'
	import CalendarIcon from '@lucide/svelte/icons/calendar'
	import { Popover } from 'bits-ui'

	import { Button } from '../button/index.js'
	import { Calendar } from '../calendar/index.js'
	import { cn } from '../cn.js'
	import { PopoverContent } from '../popover/index.js'

	let {
		value = $bindable(),
		placeholder = 'Pick a date',
		class: className,
	}: { value?: DateValue; placeholder?: string; class?: string } = $props()

	let open = $state(false)

	const formatter = new DateFormatter(navigator.language, { dateStyle: 'medium' })
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
			<Button
				{...props}
				variant="outline"
				class={cn(
					'border-input w-full min-w-52 justify-between px-3 leading-6 font-normal',
					className,
				)}
			>
				{#if value}
					<span>{formatter.format(value.toDate(getLocalTimeZone()))}</span>
				{:else}
					<span class="text-muted">{placeholder}</span>
				{/if}
				<CalendarIcon class="text-muted group-hover:text-foreground" />
			</Button>
		{/snippet}
	</Popover.Trigger>
	<PopoverContent align="center" class="rounded-xl">
		<Calendar type="single" bind:value onValueChange={() => (open = false)} />
	</PopoverContent>
</Popover.Root>
