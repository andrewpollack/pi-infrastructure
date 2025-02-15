<script lang="ts">
	import type { PageData } from './$types';
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsList from '$lib/MealsList.svelte';

	export let data: PageData;
	const { isDev, allMeals, allCalendars } = data;
</script>

<h1>
	{#if isDev}
		((DEVMODE))
	{/if}
	Meals
</h1>

<CalendarMonth monthData={allCalendars.currMonthResponse} />
<MealsList meals={allMeals} />
