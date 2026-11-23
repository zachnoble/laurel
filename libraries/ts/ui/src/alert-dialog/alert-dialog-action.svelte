<script lang="ts">
	import { AlertDialog } from 'bits-ui'

	import { Button, type ButtonProps } from '../button/index.js'

	let { variant = 'solid', color = 'red', children, ...restProps }: ButtonProps = $props()
</script>

<AlertDialog.Action>
	{#snippet child({ props })}
		<Button {...props} {...restProps} {variant} {color}>{@render children?.()}</Button>
	{/snippet}
</AlertDialog.Action>
