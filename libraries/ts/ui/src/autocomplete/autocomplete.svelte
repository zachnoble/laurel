<script lang="ts" generics="T extends Option">
	import CheckIcon from '@lucide/svelte/icons/check'
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down'
	import LoaderCircleIcon from '@lucide/svelte/icons/loader-circle'
	import SearchIcon from '@lucide/svelte/icons/search'
	import XIcon from '@lucide/svelte/icons/x'
	import { Combobox } from 'bits-ui'

	import { cn } from '../cn.js'
	import { EmptyState } from '../empty-state/index.js'
	import { floatingSurface, listItem } from '../popover/index.js'
	import { toastError } from '../toast/index.js'
	import type { Option } from '../types.js'

	type Props = {
		fetcher: (params: { signal: AbortSignal; filterText: string }) => Promise<{ items: T[] }>
		selected?: T | null
		placeholder?: string
		errorMessage?: string
		class?: string
	}

	let {
		fetcher,
		selected = $bindable(null),
		placeholder,
		errorMessage,
		class: className,
	}: Props = $props()

	let inputValue = $state(selected?.name ?? '')
	let open = $state(false)
	let items = $state<T[]>([])
	let loading = $state(false)
	let anchor = $state<HTMLElement | null>(null)

	$effect(() => {
		const filterText = inputValue
		if (selected && filterText === selected.name) {
			items = [selected]
			return
		}

		const controller = new AbortController()
		loading = true
		fetcher({ signal: controller.signal, filterText })
			.then((result) => {
				items = result.items
				loading = false
			})
			.catch((error) => {
				if (controller.signal.aborted) return
				loading = false
				toastError(error)
			})

		return () => controller.abort()
	})

	function select(id: string) {
		selected = items.find((item) => item.id === id) ?? null
		inputValue = selected?.name ?? ''
	}

	function clear() {
		selected = null
		inputValue = ''
	}
</script>

<div class={cn('w-full min-w-56', className)}>
	<Combobox.Root
		type="single"
		bind:open
		value={selected?.id ?? ''}
		onValueChange={select}
		{inputValue}
		items={items.map((item) => ({ value: item.id, label: item.name }))}
	>
		<div
			bind:this={anchor}
			class="group/combobox isolate grid grid-cols-[36px_1fr_minmax(36px,max-content)_minmax(36px,max-content)] items-center"
		>
			{#if loading}
				<LoaderCircleIcon
					class="text-muted z-10 col-start-1 row-start-1 size-4 animate-spin place-self-center"
				/>
			{:else}
				<SearchIcon
					class="text-muted z-10 col-start-1 row-start-1 size-4 place-self-center"
				/>
			{/if}

			<Combobox.Input
				{placeholder}
				aria-label={placeholder}
				aria-invalid={errorMessage ? true : undefined}
				oninput={(event) => {
					inputValue = event.currentTarget.value
					if (!inputValue) selected = null
				}}
				onfocus={() => (open = true)}
				onblur={() => {
					if (!selected && inputValue !== '') clear()
				}}
				class={cn(
					'border-input col-span-full row-start-1 w-full rounded-md border py-[calc(--spacing(1.5)-1px)] ps-8 pe-16 shadow-xs outline-hidden',
					'placeholder:text-muted text-sm/6 dark:shadow-none',
					'focus:border-ring focus:ring-ring focus:ring-1 aria-invalid:border-red-600',
				)}
			/>

			<button
				type="button"
				aria-label="Clear"
				onclick={clear}
				class={cn(
					'text-muted hover:text-foreground -col-end-2 row-start-1 -me-1 flex size-7 items-center justify-center justify-self-end rounded-md',
					inputValue ? 'visible' : 'invisible',
				)}
			>
				<XIcon class="size-4" />
			</button>

			<Combobox.Trigger
				aria-label="Show options"
				class="text-muted/75 group-hover/combobox:text-foreground -col-end-1 row-start-1 flex size-7 items-center justify-center place-self-center rounded-md"
			>
				<ChevronDownIcon class="size-4" />
			</Combobox.Trigger>
		</div>

		<Combobox.Portal>
			<Combobox.Content
				customAnchor={anchor}
				sideOffset={8}
				class={cn(
					floatingSurface,
					'max-h-72 w-(--bits-combobox-anchor-width) overflow-auto',
				)}
			>
				<Combobox.Viewport class="flex flex-col p-1">
					{#each items as item (item.id)}
						<Combobox.Item
							value={item.id}
							label={item.name}
							disabled={item.disabled}
							class={cn(listItem)}
						>
							{#snippet children({ selected })}
								<span class="w-full truncate">{item.name}</span>
								<CheckIcon
									class={cn('size-4', selected ? 'visible' : 'invisible')}
								/>
							{/snippet}
						</Combobox.Item>
					{:else}
						<EmptyState
							heading="No results found"
							description="Try searching for something else"
						/>
					{/each}
				</Combobox.Viewport>
			</Combobox.Content>
		</Combobox.Portal>
	</Combobox.Root>

	{#if errorMessage}
		<p class="mt-2 block text-sm/6 text-red-600">{errorMessage}</p>
	{/if}
</div>
