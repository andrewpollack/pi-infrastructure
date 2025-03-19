<script lang="ts">
	import type { PageData } from './$types';
	import MealsList from '$lib/MealsList.svelte';
	import Header from '$lib/Header.svelte';

	export let data: PageData;
	const { meals } = data;
</script>

<Header title="Update Meals" />

<MealsList {meals} />
