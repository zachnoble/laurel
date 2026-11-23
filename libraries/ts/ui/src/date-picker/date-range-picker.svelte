<script lang="ts">
	import { DateFormatter, getLocalTimeZone } from '@internationalized/date'
	import CalendarIcon from '@lucide/svelte/icons/calendar'
	import { type DateRange, Popover } from 'bits-ui'

	import { Button } from '../button/index.js'
	import { cn } from '../cn.js'
	import { PopoverContent } from '../popover/index.js'
	import { RangeCalendar } from '../range-calendar/index.js'

	let {
		value = $bindable(),
		placeholder = 'Pick a date range',
		class: className,
	}: { value?: DateRange; placeholder?: string; class?: string } = $props()

	let open = $state(false)

	const formatter = new DateFormatter(navigator.language, { dateStyle: 'medium' })
	const format = (date: NonNullable<DateRange['start']>) =>
		formatter.format(date.toDate(getLocalTimeZone()))
</script>

<Popover.Root bind:open>
	<Popover.Trigger>
		{#snippet child({ props })}
			<Button
				{...props}
				variant="outline"
				class={cn('border-input w-full min-w-64 px-0 font-normal', className)}
			>
				{#if value?.start && value?.end}
					<div class="grid w-full grid-cols-[1fr_16px_1fr_36px] items-center">
						<span class="min-w-fit px-3 text-sm/6">{format(value.start)}</span>
						<span aria-hidden="true" class="text-muted place-self-center">–</span>
						<span class="min-w-fit px-3 text-sm/6">{format(value.end)}</span>
						<CalendarIcon
							class="text-muted group-hover:text-foreground place-self-center"
						/>
					</div>
				{:else}
					<div class="grid w-full grid-cols-[1fr_36px] items-center">
						<span class="text-muted justify-self-start px-3">{placeholder}</span>
						<CalendarIcon
							class="text-muted group-hover:text-foreground place-self-center"
						/>
					</div>
				{/if}
			</Button>
		{/snippet}
	</Popover.Trigger>
	<PopoverContent align="center" class="rounded-xl">
		<RangeCalendar
			bind:value
			onValueChange={(range) => {
				if (range.start && range.end) open = false
			}}
		/>
	</PopoverContent>
</Popover.Root>
