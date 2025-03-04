<script lang="ts">
	import type { PageData } from './$types';
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsList from '$lib/MealsList.svelte';
	import MealsEmail from '$lib/MealsEmail.svelte';

	export let data: PageData;
	const { isDev, allMeals, allCalendars } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{#if isDev}
		((DEVMODE))
	{/if}
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
	<span>Meals</span>
	<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />
</h1>

<CalendarMonth monthData={allCalendars.currMonthResponse} />
<div style="display: flex; gap: 1rem;">
	<div style="flex: 1;">
		<MealsList meals={allMeals} />
	</div>
	<div style="flex: 1;">
		<MealsEmail meals={allMeals} />
	</div>
</div>
