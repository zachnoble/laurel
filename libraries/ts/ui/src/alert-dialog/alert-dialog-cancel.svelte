<script lang="ts">
	import { AlertDialog } from 'bits-ui'

	import { Button, type ButtonProps } from '../button/index.js'

	let { variant = 'outline', children, ...restProps }: ButtonProps = $props()
</script>

<AlertDialog.Cancel>
	{#snippet child({ props })}
		<Button {...props} {...restProps} {variant}>{@render children?.()}</Button>
	{/snippet}
</AlertDialog.Cancel>
