<script lang="ts">
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsEmail from '$lib/MealsEmail.svelte';

	const icon = '<img src="/favicon.ico" alt="Favicon" style="height: 1em;" />';

	let { data } = $props();
	const { allMeals, currMonthResponse, allEmails, allExtraItems } = data;
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html icon}
	<span>Meals</span>
	{@html icon}
</h1>

<CalendarMonth monthData={currMonthResponse} />

<div>
	<MealsEmail meals={allMeals} emails={allEmails} extraItems={allExtraItems} />
</div>
