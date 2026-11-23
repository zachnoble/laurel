<script lang="ts">
	import CheckIcon from '@lucide/svelte/icons/check'
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down'
	import XIcon from '@lucide/svelte/icons/x'
	import { Select } from 'bits-ui'

	import { cn } from '../cn.js'
	import { floatingSurface, listItem } from '../popover/index.js'
	import type { Option, OptionGroup } from '../types.js'

	type Props = {
		value?: string
		options: Option[] | OptionGroup[]
		name?: string
		placeholder?: string
		disabled?: boolean
		error?: string
		clearable?: boolean
		class?: string
		onValueChange?: (value: string) => void
	}

	let {
		value = $bindable(''),
		options,
		name,
		placeholder = 'Select an option',
		disabled,
		error,
		clearable,
		class: className,
		onValueChange,
	}: Props = $props()

	const isGrouped = (opts: Option[] | OptionGroup[]): opts is OptionGroup[] =>
		opts[0] !== undefined && 'options' in opts[0]

	const groups = $derived<OptionGroup[]>(isGrouped(options) ? options : [{ title: '', options }])
	const flatOptions = $derived(groups.flatMap((group) => group.options))
	const selected = $derived(flatOptions.find((option) => option.id === value))

	const errorId = $props.id()
</script>

<div class={cn('w-full min-w-56', className)}>
	<Select.Root
		type="single"
		bind:value
		{onValueChange}
		{name}
		{disabled}
		items={flatOptions.map((option) => ({ value: option.id, label: option.name }))}
	>
		<div class="relative">
			<Select.Trigger
				aria-label={placeholder}
				aria-invalid={error ? true : undefined}
				aria-describedby={error ? errorId : undefined}
				class={cn(
					'group border-input flex w-full cursor-pointer items-center gap-x-2 rounded-md border text-start shadow-xs outline-hidden transition',
					'py-[calc(--spacing(1.5)-1px)] ps-3.5 pe-2.5 text-sm/6',
					'hover:bg-subtle data-[state=open]:bg-subtle aria-invalid:border-red-600',
					'focus-visible:border-ring focus-visible:ring-ring focus-visible:ring-1',
					'data-disabled:cursor-not-allowed data-disabled:opacity-50',
				)}
			>
				<span
					class={cn(
						'flex-1 truncate',
						!selected && 'text-muted dark:text-white',
						clearable && value && 'me-6',
					)}
				>
					{selected?.name ?? placeholder}
				</span>
				<ChevronDownIcon class="text-muted/75 group-hover:text-foreground size-4" />
			</Select.Trigger>

			{#if clearable && value}
				<button
					type="button"
					aria-label="Clear selection"
					class="text-muted hover:text-foreground absolute inset-y-0 end-8 flex items-center"
					onclick={() => {
						value = ''
						onValueChange?.('')
					}}
				>
					<XIcon class="size-3.5" />
				</button>
			{/if}
		</div>

		<Select.Portal>
			<Select.Content
				sideOffset={8}
				class={cn(
					floatingSurface,
					'max-h-(--bits-select-content-available-height) w-(--bits-select-anchor-width) overflow-auto',
				)}
			>
				<Select.Viewport class="flex flex-col p-1 has-[[data-select-group-label]]:pt-0">
					{#each groups as group (group.title)}
						<Select.Group
							class="not-first:border-border/75 not-first:mt-1.5 not-first:border-t"
						>
							{#if group.title}
								<Select.GroupHeading
									class="bg-background text-muted sticky inset-0 z-10 truncate px-2 pt-2 text-xs/6"
								>
									{group.title}
								</Select.GroupHeading>
							{/if}
							{#each group.options as option (option.id)}
								<Select.Item
									value={option.id}
									label={option.name}
									disabled={option.disabled}
									class={cn(listItem)}
								>
									{#snippet children({ selected })}
										<div class="grid w-full">
											<span class="truncate">{option.name}</span>
											{#if option.description}
												<small class="text-muted text-xs"
													>{option.description}</small
												>
											{/if}
										</div>
										<CheckIcon
											class={cn('size-4', selected ? 'visible' : 'invisible')}
										/>
									{/snippet}
								</Select.Item>
							{/each}
						</Select.Group>
					{/each}
				</Select.Viewport>
			</Select.Content>
		</Select.Portal>
	</Select.Root>

	{#if error}
		<p id={errorId} class="mt-2 block text-sm/6 text-red-600">{error}</p>
	{/if}
</div>
