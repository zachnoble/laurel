export const modalOverlay = [
	'fixed inset-0 z-40 bg-zinc-950/25 dark:bg-zinc-950/50',
	'data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:duration-300 data-[state=open]:ease-out',
	'data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:duration-200 data-[state=closed]:ease-in',
]

export const modalPositioner = 'pointer-events-none fixed inset-0 z-50 grid justify-items-center'

export const modalContent = [
	'bg-background pointer-events-auto relative row-start-2 flex max-h-full w-full flex-col overflow-hidden',
	'text-left shadow-lg ring-1 ring-zinc-950/5 outline-hidden dark:ring-zinc-800',
	'[&:not(:has([data-ui=dialog-footer]))>[data-ui=dialog-body]]:pb-6',
	'[&:not(:has([data-ui=dialog-header]))>[data-ui=dialog-body]]:pt-6',
	'data-[state=open]:animate-in data-[state=open]:duration-200 data-[state=open]:ease-out',
	'data-[state=closed]:animate-out data-[state=closed]:duration-200 data-[state=closed]:ease-in',
]
