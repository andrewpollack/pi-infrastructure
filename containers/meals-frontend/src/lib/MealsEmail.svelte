<script lang="ts">
	import type { Meal } from '$lib/types';
	import EmailMealItem from './EmailMealItem.svelte';
	import StatusIndicator from './StatusIndicator.svelte';

	let errorMessage: string | null = null;
	let successMessage: string | null = null;
	let isLoading = false;

	export let meals: Meal[];
	export let emails: string[];

	let selectedMeals: string[] = [];
	let selectedEmails: string[] = [];

	const maxMeals = 7;
	const staticMeals: Meal[] = [
		{ Day: 0, Meal: 'Out', URL: null, Enabled: null },
		{ Day: 0, Meal: 'Leftovers', URL: null, Enabled: null }
	];
	const allMealItems: Meal[] = [...staticMeals, ...meals];
	const daysOfWeek = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
	const shortenedDaysOfWeek = ['Sun', 'Mon', 'Tue', 'Wed', 'Thur', 'Fri', 'Sat'];

	function toggleMeal(meal: string, checked: boolean) {
		if (checked) {
			// Enforce a maximum of maxMeals selections
			if (selectedMeals.length >= maxMeals) {
				alert(`error: You can only select up to ${maxMeals} meals.`);
				return;
			}
			selectedMeals = [...selectedMeals, meal];
		} else {
			selectedMeals = selectedMeals.filter((m) => m !== meal);
		}
	}

	function toggleEmail(email: string, checked: boolean) {
		if (checked) {
			selectedEmails = [...selectedEmails, email];
		} else {
			selectedEmails = selectedEmails.filter((em) => em !== email);
		}
	}

	const numColumns = 2;
	const itemsPerColumn = Math.ceil(allMealItems.length / numColumns);

	const chunkedMeals: Meal[][] = [];
	for (let i = 0; i < numColumns; i++) {
		const start = i * itemsPerColumn;
		const end = start + itemsPerColumn;
		chunkedMeals.push(allMealItems.slice(start, end));
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();
		errorMessage = null;
		successMessage = null;
		isLoading = true;
		var paddedMeals = selectedMeals;

		if (paddedMeals.length < 7) {
			// Pad selectedMeals with "Out" strings to ensure there are 7 meals
			paddedMeals = [...paddedMeals, ...Array(7 - paddedMeals.length).fill('Out')];
		}

		try {
			const res = await fetch('/api/email', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					meals: paddedMeals,
					emails: selectedEmails
				})
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending data.');
			}

			const data = await res.json();
			console.log('Response:', data);
			successMessage = 'Email sent successfully!';
		} catch (error) {
			console.error('Error sending meals and emails:', error);
			errorMessage = 'error: ';
			if (error instanceof Error) {
				errorMessage += error.message;
			} else {
				errorMessage += String(error);
			}
			alert(errorMessage);
		} finally {
			isLoading = false;
		}
	}
</script>

<h2>Email</h2>

<StatusIndicator {isLoading} {successMessage} {errorMessage} />

<form on:submit={handleSubmit}>
	<button type="submit">Send Email</button>

	<table border="1" style="margin-top: 1rem; border-collapse: collapse;">
		<thead>
			<tr>
				{#each daysOfWeek as day}
					<th>{day}</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			<tr>
				{#each [...Array(maxMeals).keys()] as i}
					<td
						style="
						max-width: 75px;
						overflow: hidden;
						text-overflow: ellipsis;
						white-space: nowrap;
						"
					>
						{#if selectedMeals[i]}
							{selectedMeals[i]}
						{:else}
							&nbsp;
						{/if}
					</td>
				{/each}
			</tr>
		</tbody>
	</table>

	<div>
		<h3>Emails</h3>
		{#each emails as email}
			<div>
				<label>
					<input
						type="checkbox"
						checked={selectedEmails.includes(email)}
						on:change={(e) => toggleEmail(email, (e.target as HTMLInputElement).checked)}
					/>
					{email}
				</label>
			</div>
		{/each}
	</div>

	<div>
		<h3>Meals</h3>
		<div style="display: flex; gap: 2rem;">
			{#each chunkedMeals as chunk}
				<div>
					{#each chunk as meal}
						<EmailMealItem
							{meal}
							isSelected={selectedMeals.includes(meal.Meal)}
							dayOfWeek={shortenedDaysOfWeek[selectedMeals.indexOf(meal.Meal)]}
							{maxMeals}
							selectedMealsCount={selectedMeals.length}
							onToggle={toggleMeal}
						/>
					{/each}
				</div>
			{/each}
		</div>
	</div>
</form>
