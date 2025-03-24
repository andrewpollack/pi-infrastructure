<script lang="ts">
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsEmail from '$lib/MealsEmail.svelte';
	import Navbar from '$lib/Navbar.svelte';

	let { data } = $props();
	const { allMeals, currMonthResponse, allEmails } = data;
</script>

<Navbar title="Meals" />

<CalendarMonth monthData={currMonthResponse} />

<div>
	<MealsEmail meals={allMeals} emails={allEmails} />
</div>
