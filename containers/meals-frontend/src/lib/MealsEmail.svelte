<script lang="ts">
	import type { Meal } from '$lib/types';

	export let meals: Meal[];

	let selectedMeals: string[] = [];
	let errorMessage: string | null = null;
	let successMessage: string | null = null;

	function toggleMeal(meal: string, checked: boolean) {
		if (checked) {
			selectedMeals = [...selectedMeals, meal];
		} else {
			selectedMeals = selectedMeals.filter((m) => m !== meal);
		}
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();
		errorMessage = null;
		successMessage = null;

		try {
			const res = await fetch('/api/email', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(selectedMeals)
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending meals.');
			}

			const data = await res.json();
			console.log('Response:', data);
			successMessage = 'Email sent successfully!';
		} catch (error) {
			console.error('Error sending meals:', error);
			errorMessage = 'error: ';
			if (error instanceof Error) {
				errorMessage += error.message;
			} else {
				errorMessage += String(error);
			}
			alert(errorMessage);
		}
	}
</script>

<h2>Email</h2>

{#if successMessage}
	<div class="success" style="color: green;">
		<p>{successMessage}</p>
	</div>
{/if}

{#if meals && meals.length > 0}
	<form on:submit={handleSubmit}>
		<button type="submit">Send Email</button>
		{#each meals as meal, index (meal.Meal + meal.Day)}
			<div class="meal-item">
				<label>
					<input
						type="checkbox"
						name="meals"
						checked={selectedMeals.includes(meal.Meal)}
						on:change={(e) => {
							const input = e.target as HTMLInputElement;
							// If checking this box would be the 6th selection, uncheck it immediately and alert.
							if (input.checked && selectedMeals.length >= 5) {
								input.checked = false;
								alert('error: You can only select up to 5 meals.');
								return;
							}
							toggleMeal(meal.Meal, input.checked);
						}}
					/>
					{meal.Meal}
				</label>
			</div>
		{/each}
	</form>
{:else}
	<p>No meals found.</p>
{/if}
