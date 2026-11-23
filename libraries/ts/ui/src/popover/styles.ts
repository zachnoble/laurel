export const floatingSurface = [
	'bg-background z-50 rounded-md shadow-lg ring-1 ring-zinc-950/10 outline-hidden dark:ring-zinc-800',
	'data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95',
	'data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95',
]

export const listItem = [
	'group flex cursor-pointer items-center gap-x-1.5 rounded-sm px-2 py-1.5 text-sm/6 outline-hidden select-none',
	'data-highlighted:bg-subtle data-disabled:opacity-50',
]
