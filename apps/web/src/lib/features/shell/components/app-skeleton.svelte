<script lang="ts">
	import { Skeleton } from '@laurel/ui/skeleton/index.js'
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left'

	import Logo from '#lib/features/shell/components/logo.svelte'

	const navLabelWidths = ['w-16', 'w-20', 'w-14', 'w-24', 'w-14', 'w-20']
	const sections = [
		{ label: 'w-20', body: 'h-9' },
		{ label: 'w-28', body: 'h-9' },
		{ label: 'w-16', body: 'h-24' },
	]
</script>

<div id="app-skeleton" class="group" aria-hidden="true">
	<div class="flex w-full justify-between group-data-[variant=auth]:hidden">
		<div
			class="bg-background fixed top-0 right-0 left-0 flex items-center justify-between px-4 py-3 md:hidden"
		>
			<Logo />
			<Skeleton class="size-8" />
		</div>

		<div class="sticky top-0 hidden h-dvh w-20 flex-col p-4.5 md:flex lg:w-60 lg:p-5.5">
			<div class="flex justify-center lg:justify-start">
				<Logo />
			</div>
			<div class="flex flex-col gap-2 py-6.5">
				{#each navLabelWidths as width, index (index)}
					<div class="flex h-12 items-center justify-center gap-3.5 p-3 lg:justify-start">
						<Skeleton class="size-5" />
						<Skeleton class="hidden h-4 lg:block {width}" />
					</div>
				{/each}
			</div>
			<Skeleton class="h-9 w-11 lg:w-full" />
		</div>

		<div class="mt-14 min-w-0 flex-1 md:mt-0">
			<div class="max-w-xl space-y-8 p-4">
				<div class="space-y-2">
					<Skeleton class="h-4 w-24" />
					<Skeleton class="h-4 w-3/4" />
				</div>
				{#each sections as section, index (index)}
					<div class="space-y-2">
						<Skeleton class="h-4 {section.label}" />
						<Skeleton class="w-full {section.body}" />
					</div>
				{/each}
			</div>
		</div>

		<div class="sticky top-0 hidden h-dvh w-80 p-5.5 md:block lg:w-105">
			<Skeleton class="mb-4 h-9 w-full" />
			<Skeleton class="h-64 w-full rounded-sm" />
		</div>
	</div>

	<div
		class="hidden min-h-dvh w-full items-center justify-center py-0 group-data-[variant=auth]:flex sm:py-16"
	>
		<div class="mx-auto w-screen sm:max-w-118.75">
			<div
				class="sm:dark:bg-subtle rounded-none border-0 p-8 sm:rounded-sm sm:border sm:p-12"
			>
				<div class="mb-8 flex flex-col items-center gap-3">
					<Skeleton class="h-7 w-40" />
					<Skeleton class="h-4 w-56" />
				</div>
				<div class="flex flex-col gap-4">
					<Skeleton class="h-9 w-full" />
					<Skeleton class="h-9 w-full" />
					<Skeleton class="h-9 w-full" />
				</div>
			</div>
		</div>
		<div class="absolute top-3 left-3 flex items-center gap-2 px-3 py-1.5 sm:top-5 sm:left-5">
			<ChevronLeftIcon class="text-muted size-4" />
			<Skeleton class="h-4 w-10" />
		</div>
	</div>
</div>
