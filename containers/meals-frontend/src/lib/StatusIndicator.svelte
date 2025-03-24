<script lang="ts">
	export let isLoading: boolean = false;
	export let loadingMessage: string = 'loading...';
	export let loadingColor: string = 'blue';

	export let successMessage: string | null = null;
	export let errorMessage: string | null = null;
</script>

{#if isLoading}
	<div style="color: {loadingColor};">
		<p>{loadingMessage}</p>
	</div>
{/if}

{#if successMessage}
	<div style="color: green;">
		<p>{successMessage}</p>
	</div>
{/if}

{#if errorMessage}
	<div style="color: red;">
		<p>{errorMessage}</p>
	</div>
{/if}
