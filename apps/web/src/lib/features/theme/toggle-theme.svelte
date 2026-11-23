<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import MoonIcon from '@lucide/svelte/icons/moon'
	import SunIcon from '@lucide/svelte/icons/sun'
	import { mode, toggleMode } from 'mode-watcher'
</script>

<Button onclick={toggleMode} variant="outline" iconOnly aria-label="Toggle theme">
	{#if mode.current === 'dark'}
		<SunIcon />
	{:else}
		<MoonIcon />
	{/if}
</Button>
