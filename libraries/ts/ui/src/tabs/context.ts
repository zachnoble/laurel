import { createContext } from 'svelte'

export type TabsVariant = 'underline' | 'pills' | 'segment'
export type TabsOrientation = 'horizontal' | 'vertical'

export const [getTabsContext, setTabsContext] = createContext<{
	readonly variant: TabsVariant
	readonly orientation: TabsOrientation
}>()
