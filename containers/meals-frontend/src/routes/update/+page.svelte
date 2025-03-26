<script lang="ts">
	import MealsList from '$lib/MealsList.svelte';

	const icon = '<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />';

	let { data } = $props();
	const { meals } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html icon}
	<span>Update Meals</span>
	{@html icon}
</h1>

<MealsList {meals} />
