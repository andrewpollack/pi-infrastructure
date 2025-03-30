<script lang="ts">
	import { StatusType } from '$lib/types';
	import { goto } from '$app/navigation';
	import StatusIndicator from '$lib/StatusIndicator.svelte';

	let message = '';
	let statusType = StatusType.SUCCESS;
	let password = '';

	async function handleSubmit(event: Event) {
		event.preventDefault();
		message = 'Logging in...';
		statusType = StatusType.LOADING;

		const res = await fetch('/api/login', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				password: password
			})
		});

		if (!res.ok) {
			if (res.status === 401) {
				message = 'Invalid password. Please try again.';
				statusType = StatusType.ERROR;
				return;
			}
			const errorData = await res.json();
			throw new Error(errorData.error || 'An error occurred while sending data.');
		}

		message = 'Logged in successfully!';
		statusType = StatusType.SUCCESS;
		goto('/');
	}
</script>

<h1>Login</h1>

{#if message}
	<StatusIndicator {message} type={statusType} />
{/if}

<form method="post" on:submit={handleSubmit}>
	<label for="password">Password</label>
	<input bind:value={password} type="password" required />
	<button type="submit">Login</button>
</form>
