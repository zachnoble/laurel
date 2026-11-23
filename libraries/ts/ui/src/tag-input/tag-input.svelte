<script lang="ts">
	import XIcon from '@lucide/svelte/icons/x'
	import type { HTMLInputAttributes } from 'svelte/elements'

	import { cn } from '../cn.js'
	import { Label } from '../label/index.js'

	type Props = Omit<HTMLInputAttributes, 'value' | 'name'> & {
		tags?: string[]
		label?: string
		name?: string
		onTagAdd?: (tag: string) => void
		onTagRemove?: (tag: string) => void
	}

	const uid = $props.id()

	let {
		tags = $bindable([]),
		label,
		name,
		class: className,
		id = uid,
		onTagAdd,
		onTagRemove,
		...restProps
	}: Props = $props()

	let inputValue = $state('')

	function addTags() {
		for (const raw of inputValue.split(/[,;]/)) {
			const tag = raw.trim().replace(/\s\s+/g, ' ')
			const exists = tags.some((t) => t.toLocaleLowerCase() === tag.toLocaleLowerCase())

			if (tag && !exists) {
				tags = [...tags, tag]
				onTagAdd?.(tag)
			}
		}

		inputValue = ''
	}

	function removeTag(index: number) {
		const removed = tags[index]
		tags = tags.filter((_, i) => i !== index)
		if (removed !== undefined) onTagRemove?.(removed)
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' || event.key === ',' || event.key === ';') {
			event.preventDefault()
			addTags()
		} else if (event.key === 'Backspace' && inputValue === '' && tags.length > 0) {
			removeTag(tags.length - 1)
		}
	}
</script>

<div class={cn('flex w-full flex-col gap-2', className)}>
	{#if label}
		<Label for={id}>{label}</Label>
	{/if}

	<div
		class={cn(
			'flex min-h-9 items-center rounded-md border',
			'has-[input:focus]:border-ring has-[input:focus]:ring-ring has-[input:focus]:ring-1',
			'has-[input[aria-invalid=true]]:border-red-600',
		)}
	>
		<div class="inline-flex flex-1 flex-wrap items-center gap-1 px-2 py-1.25">
			<ul class="contents" aria-label={label ? `${label} list` : 'Tags'}>
				{#each tags as tag, index (tag)}
					<li
						class="flex max-w-fit items-center gap-x-1 rounded-md bg-zinc-500/10 px-2 py-0.5 text-xs/5 font-medium text-zinc-700 dark:bg-white/10 dark:text-zinc-300"
					>
						{tag}
						<button
							type="button"
							aria-label={`Remove ${tag}`}
							onclick={() => removeTag(index)}
							class="focus-visible:outline-ring flex cursor-pointer items-center justify-center rounded-full p-0.5 outline-0 transition-[background-color] hover:bg-black/10 focus-visible:outline-2 focus-visible:outline-offset-2 active:bg-black/20 dark:hover:bg-white/10 dark:active:bg-white/20"
						>
							<XIcon class="size-3" />
						</button>
					</li>
				{/each}
			</ul>

			<input
				{id}
				bind:value={inputValue}
				{onkeydown}
				class="min-w-0 flex-1 border-0 bg-transparent px-0.5 text-sm/6 outline-hidden"
				{...restProps}
			>
		</div>
	</div>

	{#if name}
		<input type="hidden" {name} value={tags.join(',')}>
	{/if}
</div>
