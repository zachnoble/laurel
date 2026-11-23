<script lang="ts">
	import { Dialog } from 'bits-ui'

	import { Button, type ButtonProps } from '../button/index.js'

	let { variant = 'outline', children, ...restProps }: ButtonProps = $props()
</script>

<Dialog.Close>
	{#snippet child({ props })}
		<Button {...props} {...restProps} {variant}>{@render children?.()}</Button>
	{/snippet}
</Dialog.Close>
