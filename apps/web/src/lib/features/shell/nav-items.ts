import CalendarIcon from '@lucide/svelte/icons/calendar'
import HouseIcon from '@lucide/svelte/icons/house'
import MailIcon from '@lucide/svelte/icons/mail'
import MessageSquareIcon from '@lucide/svelte/icons/message-square'
import SearchIcon from '@lucide/svelte/icons/search'
import SettingsIcon from '@lucide/svelte/icons/settings'
import type { Component } from 'svelte'

export type NavItem = {
	label: string
	href: string
	icon: Component<{ class?: string }>
}

export const navItems: NavItem[] = [
	{ label: 'Home', href: '/', icon: HouseIcon },
	{ label: 'Search', href: '/search', icon: SearchIcon },
	{ label: 'Mail', href: '/mail', icon: MailIcon },
	{ label: 'Calendar', href: '/calendar', icon: CalendarIcon },
	{ label: 'Chat', href: '/chat', icon: MessageSquareIcon },
	{ label: 'Settings', href: '/settings', icon: SettingsIcon },
]
