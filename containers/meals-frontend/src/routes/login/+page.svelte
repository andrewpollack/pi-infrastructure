<!-- +page.svelte -->
<script lang="ts">
	import type { PageProps } from './$types';
	import { StatusType } from '$lib/types';
	import StatusIndicator from '$lib/StatusIndicator.svelte';

	let message = $state('');

	let { form }: PageProps = $props();
</script>

<h1>Login</h1>

{#if message}
	<StatusIndicator {message} type={StatusType.LOADING} />
{/if}

<form
	method="post"
	onsubmit={() => {
		message = 'Loading';
	}}
>
	{#if form?.message}
		<div class="error" style="color: red;">
			{form.message}
		</div>
	{/if}
	<label for="password">Password</label>
	<input id="password" name="password" type="password" required />

	<button type="submit">Login</button>
</form>
