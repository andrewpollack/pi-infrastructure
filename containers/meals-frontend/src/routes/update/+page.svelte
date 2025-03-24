<script lang="ts">
	import MealsList from '$lib/MealsList.svelte';
	import Navbar from '$lib/Navbar.svelte';

	let { data } = $props();
	const { meals } = data;
</script>

<Navbar title="Update Meals" />

<MealsList {meals} />
