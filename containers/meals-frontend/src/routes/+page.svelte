<script lang="ts">
	import type { PageData } from './$types';
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsEmail from '$lib/MealsEmail.svelte';
	import Header from '$lib/Header.svelte';

	export let data: PageData;
	const { allMeals, allCalendars, allEmails } = data;
</script>

<Header title="Meals" />

<CalendarMonth monthData={allCalendars.currMonthResponse} />

<div>
	<MealsEmail meals={allMeals} emails={allEmails} />
</div>
