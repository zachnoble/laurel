import type { Handle } from '@sveltejs/kit/hooks'
import { createInitialModeExpression } from 'mode-watcher'
import { render } from 'svelte/server'
import { API_URL } from '$app/env/public'

import AppSkeleton from '#lib/features/shell/components/app-skeleton.svelte'

const themeScript = `<script>${createInitialModeExpression()}</script>`

const apiPreconnect = `<link rel="preconnect" href="${new URL(API_URL).origin}" crossorigin="use-credentials">`

const devStylesheet = import.meta.env.DEV
	? '<link rel="stylesheet" href="/src/routes/layout.css?direct">'
	: ''

const skeleton = `${render(AppSkeleton).body}<script>
	if (/^\\/login(\\/|$)/.test(location.pathname)) {
		document.getElementById('app-skeleton').dataset.variant = 'auth'
	}
</script>`

export const handle: Handle = ({ event, resolve }) =>
	resolve(event, {
		transformPageChunk: ({ html }) =>
			html
				.replace('%api.preconnect%', apiPreconnect)
				.replace('%theme.script%', themeScript + devStylesheet)
				.replace('%app.skeleton%', skeleton),
	})
