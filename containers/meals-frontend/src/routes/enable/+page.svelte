<script lang="ts">
	import EnableMeals from '$lib/EnableMeals.svelte';
	import { Icon } from '$lib/const';

	let { data } = $props();
	const { meals } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html Icon}
	<span>Enable Meals</span>
	{@html Icon}
</h1>

<EnableMeals {meals} />
