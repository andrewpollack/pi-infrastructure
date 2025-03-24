<script lang="ts">
	import { page } from '$app/state';
	import Navbar from '$lib/Navbar.svelte';
</script>

<Navbar title="" />

<div class="error-container">
	<h1>Error {page?.status}</h1>
	<p>{page?.error?.message || 'An unexpected error occurred.'}</p>
</div>

<style>
	h1 {
		color: #b7b1f2;
	}
</style>
