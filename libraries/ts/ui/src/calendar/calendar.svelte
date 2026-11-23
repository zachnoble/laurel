<script lang="ts">
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left'
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right'
	import { Calendar, type WithoutChildrenOrChild } from 'bits-ui'

	import { cn } from '../cn.js'
	import {
		calendarGrid,
		calendarHeadCell,
		calendarHeader,
		calendarHeading,
		calendarNavButton,
		calendarOutsideMonth,
		calendarRoot,
		calendarToday,
	} from './styles'

	let {
		ref = $bindable(null),
		value = $bindable(),
		placeholder = $bindable(),
		class: className,
		weekdayFormat = 'short',
		...restProps
	}: WithoutChildrenOrChild<Calendar.RootProps> = $props()
</script>

<Calendar.Root
	bind:ref
	bind:value={value as never}
	bind:placeholder
	{weekdayFormat}
	class={cn(calendarRoot, className)}
	{...restProps}
>
	{#snippet children({ months, weekdays })}
		<Calendar.Header class={calendarHeader}>
			<Calendar.Heading class={calendarHeading} />
			<Calendar.PrevButton class={calendarNavButton} aria-label="Previous">
				<ChevronLeftIcon />
			</Calendar.PrevButton>
			<Calendar.NextButton class={calendarNavButton} aria-label="Next">
				<ChevronRightIcon />
			</Calendar.NextButton>
		</Calendar.Header>

		{#each months as month (month.value.toString())}
			<Calendar.Grid class={calendarGrid}>
				<Calendar.GridHead>
					<Calendar.GridRow>
						{#each weekdays as day (day)}
							<Calendar.HeadCell class={calendarHeadCell}>{day}</Calendar.HeadCell>
						{/each}
					</Calendar.GridRow>
				</Calendar.GridHead>
				<Calendar.GridBody>
					{#each month.weeks as week (week[0]?.toString())}
						<Calendar.GridRow>
							{#each week as date (date.toString())}
								<Calendar.Cell {date} month={month.value} class="p-0">
									<Calendar.Day
										class={cn(
											'relative flex size-10 cursor-pointer items-center justify-center rounded-lg text-sm outline-hidden',
											'hover:bg-subtle focus-visible:outline-ring focus-visible:outline-2',
											'data-selected:bg-accent data-selected:text-white data-selected:focus-visible:outline-offset-1',
											'data-disabled:opacity-50 data-unavailable:text-red-600 data-unavailable:line-through',
											calendarToday,
											calendarOutsideMonth,
										)}
									/>
								</Calendar.Cell>
							{/each}
						</Calendar.GridRow>
					{/each}
				</Calendar.GridBody>
			</Calendar.Grid>
		{/each}
	{/snippet}
</Calendar.Root>
