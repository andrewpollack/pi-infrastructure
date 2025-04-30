<script lang="ts">
	import ExtraItems from '$lib/ExtraItems.svelte';

	let { data } = $props();
	const { extraItems, aisles } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span> Extra Items </span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<ExtraItems {extraItems} aisles={aisles} />
