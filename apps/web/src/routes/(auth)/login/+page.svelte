<script lang="ts">
	import { page } from '$app/state'

	import ConfirmLoginForm from './confirm-login-form.svelte'
	import InitiateLoginForm from './initiate-login-form.svelte'

	const email = $derived(page.url.searchParams.get('email'))
</script>

<svelte:head>
	<title>Login</title>
	<meta name="description" content="Login to your account">
</svelte:head>

{#if email}
	<ConfirmLoginForm {email} />
{:else}
	<InitiateLoginForm />
{/if}
