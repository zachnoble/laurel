import { createContext } from 'svelte'

export type CalloutContext = {
	headingId: string
	descriptionId: string
}

export const [getCalloutContext, setCalloutContext] = createContext<CalloutContext>()
