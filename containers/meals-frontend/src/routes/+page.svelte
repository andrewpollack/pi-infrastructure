<script lang="ts">
	import CalendarMonth from '$lib/CalendarMonth.svelte';
	import MealsEmail from '$lib/MealsEmail.svelte';
	import { Icon } from '$lib/const';

	let { data } = $props();
	const { allMeals, currMonthResponse, allEmails, allExtraItems, selectedYear, selectedMonth } =
		data;

	let year = $state(selectedYear);
	let month = $state(selectedMonth);
	let calendarData = $state(currMonthResponse);

	async function fetchCalendar() {
		const res = await fetch(`/api/calendar?year=${year}&month=${month}`);
		if (res.ok) {
			const data = await res.json();
			calendarData = data.currMonthResponse;
		} else {
			console.error('Failed to load calendar');
		}
	}

	function goToPrevMonth() {
		month -= 1;
		if (month < 1) {
			month = 12;
			year -= 1;
		}
		fetchCalendar();
	}

	function goToNextMonth() {
		month += 1;
		if (month > 12) {
			month = 1;
			year += 1;
		}
		fetchCalendar();
	}
</script>

<h1 style="display: flex; align-items: center; justify-content: flex-start; gap: 1rem;">
	{@html Icon}
	<span> Home </span>
	{@html Icon}
</h1>

<div style="align-items: center; justify-content: space-between; margin-bottom: 1rem;">
	<button onclick={goToPrevMonth}>Previous Month</button>
	<button onclick={goToNextMonth}>Next Month</button>
</div>

<CalendarMonth monthData={calendarData} />

<div>
	<MealsEmail meals={allMeals} emails={allEmails} extraItems={allExtraItems} />
</div>
