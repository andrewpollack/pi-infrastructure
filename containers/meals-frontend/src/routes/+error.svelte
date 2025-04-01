<script lang="ts">
	import { page } from '$app/state';
	import { Colors } from '$lib/const';
</script>

<div class="error-container" style="--primary-color: {Colors.primary};">
	<h1>Error {page?.status}</h1>
	<p>{page?.error?.message || 'An unexpected error occurred.'}</p>
</div>

<style>
	h1 {
		color: var(--primary-color);
	}
</style>
