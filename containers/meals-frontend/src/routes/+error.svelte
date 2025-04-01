<script lang="ts">
	import { page } from '$app/state';
	import { Color } from '$lib/const';
</script>

<div class="error-container" style="--primary-color: {Color.primary};">
	<h1>Error {page?.status}</h1>
	<p>{page?.error?.message || 'An unexpected error occurred.'}</p>
</div>

<style>
	h1 {
		color: var(--primary-color);
	}
</style>
