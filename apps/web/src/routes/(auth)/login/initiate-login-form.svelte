<script lang="ts">
	import { Button } from '@laurel/ui/button/index.js'
	import { Input } from '@laurel/ui/input/index.js'
	import { z } from 'zod/mini'
	import { goto } from '$app/navigation'
	import { resolve } from '$app/paths'

	import { createForm } from '#lib/form.js'
	import { authClient } from '#lib/rpc-clients.js'

	const schema = z.object({
		email: z.email('Please enter a valid email address'),
	})

	const { form, errors, enhance, submitting } = createForm(schema, async ({ email }) => {
		await authClient.sendOtp({ email })
		await goto(resolve(`login?email=${encodeURIComponent(email)}`))
	})
</script>

<div class="mb-8 text-center">
	<h1 class="text-2xl font-bold">Get Started</h1>
	<p class="text-muted mt-2 text-sm">Enter your email for a sign-in code</p>
</div>

<form method="POST" use:enhance class="flex flex-col gap-4" novalidate>
	<div>
		<Input
			name="email"
			type="email"
			autocomplete="email"
			placeholder="Enter your email"
			bind:value={$form.email}
			errorMessage={$errors.email?.[0]}
		/>
	</div>

	<Button type="submit" pending={$submitting}>Continue with email</Button>
</form>
