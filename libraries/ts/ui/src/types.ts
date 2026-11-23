export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null }

export type Option = {
	id: string
	name: string
	description?: string
	disabled?: boolean
}

export type OptionGroup = {
	title: string
	options: Option[]
}
