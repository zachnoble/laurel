<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import { Input } from '@laurel/ui/input/index.js'
	import MailIcon from '@lucide/svelte/icons/mail'
	import { z } from 'zod/mini'
	import { goto } from '$app/navigation'
	import { resolve } from '$app/paths'

	import { createForm } from '#lib/form.js'
	import { authClient } from '#lib/rpc-clients.js'

	let { email }: { email: string } = $props()

	const schema = z.object({
		code: z
			.string()
			.check(
				z.minLength(1, 'Please enter your verification code'),
				z.length(6, 'Code must be 6 digits'),
			),
	})

	const { form, errors, enhance, submitting } = createForm(schema, async ({ code }) => {
		await authClient.verifyOtp({ email, otp: code })
		await goto(resolve(''), { invalidateAll: true })
	})
</script>

<div class="flex w-full flex-col items-center justify-center">
	<MailIcon class="size-9" />

	<h1 class="py-2 text-xl font-bold">Check your email</h1>

	<p class="text-muted pb-4 text-center text-sm">We sent a code to {email}</p>

	<form method="POST" use:enhance class="flex w-full flex-col gap-4" novalidate>
		<div>
			<Input
				name="code"
				inputmode="numeric"
				autocomplete="one-time-code"
				placeholder="Enter verification code"
				bind:value={$form.code}
				errorMessage={$errors.code?.[0]}
			/>
		</div>

		<Button type="submit" pending={$submitting}>Verify your email address</Button>
	</form>

	<div class="mt-4">
		<Button href={resolve('login')} variant="link">Use a different email address</Button>
	</div>
</div>
