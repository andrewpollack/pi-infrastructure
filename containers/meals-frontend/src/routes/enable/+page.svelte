<script lang="ts">
	import EnableMeals from '$lib/EnableMeals.svelte';

	let { data } = $props();
	const { meals } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span>Enable Meals</span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<EnableMeals {meals} />
