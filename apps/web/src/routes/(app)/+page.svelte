<script lang="ts">
	import type { DateValue } from '@internationalized/date'
	import {
		AlertDialogAction,
		AlertDialogCancel,
		AlertDialogContent,
		AlertDialogHeader,
		AlertDialogRoot,
		AlertDialogTrigger,
	} from '@laurel/ui/alert-dialog/index.js'
	import { Autocomplete } from '@laurel/ui/autocomplete/index.js'
	import { Badge } from '@laurel/ui/badge/index.js'
	import { Breadcrumb, Breadcrumbs } from '@laurel/ui/breadcrumbs/index.js'
	import { Button } from '@laurel/ui/button/index.js'
	import {
		Callout,
		CalloutDescription,
		CalloutHeading,
		CalloutIcon,
	} from '@laurel/ui/callout/index.js'
	import { Checkbox, CheckboxGroup } from '@laurel/ui/checkbox/index.js'
	import { DatePicker, DateRangePicker } from '@laurel/ui/date-picker/index.js'
	import {
		DialogBody,
		DialogClose,
		DialogContent,
		DialogFooter,
		DialogHeader,
		DialogRoot,
		DialogTrigger,
	} from '@laurel/ui/dialog/index.js'
	import { Select } from '@laurel/ui/select/index.js'
	import { Skeleton } from '@laurel/ui/skeleton/index.js'
	import {
		Table,
		TableBody,
		TableCell,
		TableHead,
		TableHeader,
		TableRow,
	} from '@laurel/ui/table/index.js'
	import { Tab, TabList, TabPanel, Tabs } from '@laurel/ui/tabs/index.js'
	import { TagInput } from '@laurel/ui/tag-input/index.js'
	import type { Option } from '@laurel/ui/types.js'
	import InfoIcon from '@lucide/svelte/icons/info'
	import type { DateRange } from 'bits-ui'
	import { SvelteSet } from 'svelte/reactivity'

	import ToggleTheme from '#lib/features/theme/toggle-theme.svelte'

	const FRUITS: Option[] = [
		{ id: 'apple', name: 'Apple' },
		{ id: 'banana', name: 'Banana' },
		{ id: 'cherry', name: 'Cherry' },
		{ id: 'date', name: 'Date' },
		{ id: 'elderberry', name: 'Elderberry' },
	]

	const SIZE_OPTIONS: Option[] = [
		{ id: 'sm', name: 'Small' },
		{ id: 'md', name: 'Medium' },
		{ id: 'lg', name: 'Large' },
		{ id: 'xl', name: 'Extra Large' },
	]

	const PEOPLE = [
		{ id: 1, name: 'Alice Johnson', role: 'Engineer', status: 'Inactive' },
		{ id: 2, name: 'Bob Smith', role: 'Designer', status: 'Active' },
		{ id: 3, name: 'Carol White', role: 'Manager', status: 'Inactive' },
		{ id: 4, name: 'Dan Brown', role: 'Engineer', status: 'Active' },
	]

	const fruitFetcher = async ({ filterText }: { signal: AbortSignal; filterText: string }) => ({
		items: FRUITS.filter((f) => f.name.toLowerCase().includes(filterText.toLowerCase())),
	})

	let selectedFruit = $state<Option | null>(null)
	let size = $state('')
	let features = $state<string[]>([])
	let date = $state<DateValue>()
	let dateRange = $state<DateRange>()
	let tags = $state(['Svelte'])
	let alertOpen = $state(false)

	const selectedPeople = new SvelteSet<number>()
	const allSelected = $derived(selectedPeople.size === PEOPLE.length)

	function togglePerson(id: number) {
		if (selectedPeople.has(id)) selectedPeople.delete(id)
		else selectedPeople.add(id)
	}

	function toggleAllPeople() {
		if (allSelected) selectedPeople.clear()
		else for (const person of PEOPLE) selectedPeople.add(person.id)
	}
</script>

<svelte:head>
	<title>App Name</title>
	<meta name="description" content="App Description">
</svelte:head>

<div class="max-w-xl space-y-8 p-4">
	<section class="space-y-2">
		<p class="text-sm font-medium">Breadcrumbs</p>
		<Breadcrumbs>
			<Breadcrumb href="/">Home</Breadcrumb>
			<Breadcrumb href="/settings">Settings</Breadcrumb>
			<Breadcrumb href="#" current>Profile</Breadcrumb>
		</Breadcrumbs>
	</section>

	<ToggleTheme />

	<section class="space-y-2">
		<p class="text-sm font-medium">Skeleton</p>
		<div class="space-y-2">
			<Skeleton class="h-4 w-3/4" />
			<Skeleton class="h-4 w-1/2" />
			<Skeleton class="h-8 w-full" />
		</div>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Autocomplete</p>
		<Autocomplete
			fetcher={fruitFetcher}
			bind:selected={selectedFruit}
			placeholder="Search fruits..."
		/>
		{#if selectedFruit}
			<p class="text-muted text-sm">Selected: {selectedFruit.name}</p>
		{/if}
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Select</p>
		<Select
			name="size"
			bind:value={size}
			options={SIZE_OPTIONS}
			placeholder="Pick a size"
			clearable
		/>
		{#if size}
			<p class="text-muted text-sm">Selected: {size}</p>
		{/if}
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Checkboxes</p>
		<CheckboxGroup bind:value={features} aria-label="Features">
			<Checkbox value="dark-mode">Dark mode</Checkbox>
			<Checkbox value="notifications">Notifications</Checkbox>
			<Checkbox value="analytics">Analytics</Checkbox>
		</CheckboxGroup>
		{#if features.length > 0}
			<p class="text-muted text-sm">Enabled: {features.join(', ')}</p>
		{/if}
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Date Picker</p>
		<DatePicker bind:value={date} placeholder="Pick a date" />
		{#if date}
			<p class="text-muted text-sm">Selected: {date.toString()}</p>
		{/if}
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Date Range Picker</p>
		<DateRangePicker bind:value={dateRange} placeholder="Pick a date range" />
		{#if dateRange?.start && dateRange?.end}
			<p class="text-muted text-sm">
				{dateRange.start.toString()} – {dateRange.end.toString()}
			</p>
		{/if}
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Tag Input</p>
		<TagInput label="Tags" bind:tags />
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Badges</p>
		<div class="flex gap-3">
			<Badge color="amber">Banana</Badge>
			<Badge color="red">Apple</Badge>
			<Badge color="indigo">Grape</Badge>
			<Badge color="rose">Cherry</Badge>
			<Badge color="green">Avocado</Badge>
		</div>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Dialog</p>
		<DialogRoot>
			<DialogTrigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline">Open Dialog</Button>
				{/snippet}
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>Edit Profile</DialogHeader>
				<DialogBody>
					<p class="text-sm">
						Make changes to your profile here. Click save when you're done.
					</p>
				</DialogBody>
				<DialogFooter>
					<DialogClose>Cancel</DialogClose>
					<DialogClose variant="solid" color="accent">Save changes</DialogClose>
				</DialogFooter>
			</DialogContent>
		</DialogRoot>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Alert Dialog</p>
		<AlertDialogRoot bind:open={alertOpen}>
			<AlertDialogTrigger>
				{#snippet child({ props })}
					<Button {...props} variant="outline" color="red">Delete item</Button>
				{/snippet}
			</AlertDialogTrigger>
			<AlertDialogContent>
				<AlertDialogHeader>Are you absolutely sure?</AlertDialogHeader>
				<DialogBody>
					<p class="text-sm">
						This action cannot be undone. This will permanently delete the item and
						remove it from our servers.
					</p>
				</DialogBody>
				<DialogFooter>
					<AlertDialogCancel>Cancel</AlertDialogCancel>
					<AlertDialogAction onclick={() => (alertOpen = false)}
						>Delete</AlertDialogAction
					>
				</DialogFooter>
			</AlertDialogContent>
		</AlertDialogRoot>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Callout</p>
		<div class="space-y-3">
			<Callout color="zinc">
				<CalloutIcon><InfoIcon /></CalloutIcon>
				<CalloutHeading>Heads up</CalloutHeading>
				<CalloutDescription
					>You can add components to your app using the CLI.</CalloutDescription
				>
			</Callout>
			<Callout color="blue">
				<CalloutHeading>New feature</CalloutHeading>
				<CalloutDescription>
					Real-time collaboration is now available on all plans.
				</CalloutDescription>
			</Callout>
			<Callout color="red">
				<CalloutHeading>Destructive action</CalloutHeading>
				<CalloutDescription>This cannot be undone once confirmed.</CalloutDescription>
			</Callout>
		</div>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Tabs</p>
		<div class="space-y-6">
			<Tabs value="overview">
				<TabList aria-label="Underline tabs">
					<Tab value="overview">Overview</Tab>
					<Tab value="analytics">Analytics</Tab>
					<Tab value="settings">Settings</Tab>
				</TabList>
				<TabPanel value="overview">Overview content</TabPanel>
				<TabPanel value="analytics">Analytics content</TabPanel>
				<TabPanel value="settings">Settings content</TabPanel>
			</Tabs>
			<Tabs value="week" variant="segment">
				<TabList aria-label="Segment tabs">
					<Tab value="week">Week</Tab>
					<Tab value="month">Month</Tab>
					<Tab value="year">Year</Tab>
				</TabList>
				<TabPanel value="week">Weekly view</TabPanel>
				<TabPanel value="month">Monthly view</TabPanel>
				<TabPanel value="year">Yearly view</TabPanel>
			</Tabs>
		</div>
	</section>

	<section class="space-y-2">
		<p class="text-sm font-medium">Table</p>
		<Table aria-label="People">
			<TableHeader>
				<TableRow>
					<TableHead class="w-9">
						<Checkbox
							aria-label="Select all"
							checked={allSelected}
							indeterminate={selectedPeople.size > 0 && !allSelected}
							onCheckedChange={toggleAllPeople}
						/>
					</TableHead>
					<TableHead>Name</TableHead>
					<TableHead>Role</TableHead>
					<TableHead>Status</TableHead>
				</TableRow>
			</TableHeader>
			<TableBody>
				{#each PEOPLE as person (person.id)}
					<TableRow
						selected={selectedPeople.has(person.id)}
						class="cursor-pointer"
						onclick={() => togglePerson(person.id)}
					>
						<TableCell>
							<Checkbox
								aria-label={`Select ${person.name}`}
								checked={selectedPeople.has(person.id)}
								onclick={(event) => event.stopPropagation()}
								onCheckedChange={() => togglePerson(person.id)}
							/>
						</TableCell>
						<TableCell>{person.name}</TableCell>
						<TableCell>{person.role}</TableCell>
						<TableCell>{person.status}</TableCell>
					</TableRow>
				{/each}
			</TableBody>
		</Table>
	</section>
</div>
