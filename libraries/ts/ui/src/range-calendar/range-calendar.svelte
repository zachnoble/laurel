<script lang="ts">
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left'
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right'
	import { RangeCalendar, type WithoutChildrenOrChild } from 'bits-ui'

	import {
		calendarGrid,
		calendarHeadCell,
		calendarHeader,
		calendarHeading,
		calendarNavButton,
		calendarOutsideMonth,
		calendarRoot,
		calendarToday,
	} from '../calendar/styles.js'
	import { cn } from '../cn.js'

	let {
		ref = $bindable(null),
		value = $bindable(),
		placeholder = $bindable(),
		class: className,
		weekdayFormat = 'short',
		...restProps
	}: WithoutChildrenOrChild<RangeCalendar.RootProps> = $props()
</script>

<RangeCalendar.Root
	bind:ref
	bind:value
	bind:placeholder
	{weekdayFormat}
	class={cn(calendarRoot, className)}
	{...restProps}
>
	{#snippet children({ months, weekdays })}
		<RangeCalendar.Header class={calendarHeader}>
			<RangeCalendar.Heading class={calendarHeading} />
			<RangeCalendar.PrevButton class={calendarNavButton} aria-label="Previous">
				<ChevronLeftIcon />
			</RangeCalendar.PrevButton>
			<RangeCalendar.NextButton class={calendarNavButton} aria-label="Next">
				<ChevronRightIcon />
			</RangeCalendar.NextButton>
		</RangeCalendar.Header>

		{#each months as month (month.value.toString())}
			<RangeCalendar.Grid class={calendarGrid}>
				<RangeCalendar.GridHead>
					<RangeCalendar.GridRow>
						{#each weekdays as day (day)}
							<RangeCalendar.HeadCell class={calendarHeadCell}
								>{day}</RangeCalendar.HeadCell
							>
						{/each}
					</RangeCalendar.GridRow>
				</RangeCalendar.GridHead>
				<RangeCalendar.GridBody>
					{#each month.weeks as week (week[0]?.toString())}
						<RangeCalendar.GridRow>
							{#each week as date (date.toString())}
								<RangeCalendar.Cell {date} month={month.value} class="p-0">
									<RangeCalendar.Day
										class={cn(
											'relative flex size-10 cursor-pointer items-center justify-center text-sm outline-hidden',
											'[td:first-child_&]:rounded-s-lg [td:last-child_&]:rounded-e-lg',
											'hover:bg-subtle focus-visible:outline-ring hover:rounded-lg focus-visible:rounded-lg focus-visible:outline-2',
											'data-selected:bg-accent/7 dark:data-selected:bg-accent/35',
											'data-selection-end:rounded-lg data-selection-start:rounded-lg',
											'data-selection-start:bg-accent! data-selection-end:bg-accent!',
											'data-selection-end:text-white data-selection-start:text-white',
											'data-highlighted:bg-accent/7 dark:data-highlighted:bg-accent/35',
											'data-disabled:opacity-50 data-unavailable:text-red-600 data-unavailable:line-through',
											calendarToday,
											calendarOutsideMonth,
										)}
									/>
								</RangeCalendar.Cell>
							{/each}
						</RangeCalendar.GridRow>
					{/each}
				</RangeCalendar.GridBody>
			</RangeCalendar.Grid>
		{/each}
	{/snippet}
</RangeCalendar.Root>
