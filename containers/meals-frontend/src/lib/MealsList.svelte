<script lang="ts">
	import type { Meal } from '$lib/types';
	import StatusIndicator from './StatusIndicator.svelte';

	export let meals: Meal[];

	let errorMessage: string | null = null;
	let successMessage: string | null = null;
	let isLoading = false;

	let localMeals = meals.map((m) => ({ ...m }));

	function toggleMeal(index: number) {
		localMeals[index].Enabled = !localMeals[index].Enabled;
	}

	async function updateMeals() {
		errorMessage = null;
		successMessage = null;
		isLoading = true;

		const updates = localMeals
			.filter((meal, index) => meal.Enabled !== meals[index].Enabled)
			.map((m) => ({ name: m.Meal, disabled: !m.Enabled }));

		console.log('Meal updates:', updates);
		try {
			const res = await fetch('/api/update', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(updates)
			});
			const data = await res.json();
			console.log('Update response:', data);
			successMessage = 'Meal status updated!';
		} catch (error) {
			errorMessage = 'error updating meals: ';
			errorMessage += error instanceof Error ? error.message : String(error);
			alert(errorMessage);
		} finally {
			isLoading = false;
		}
	}
</script>

<h2>Meal Status</h2>

<StatusIndicator {isLoading} {successMessage} {errorMessage} />

{#if localMeals && localMeals.length > 0}
	<form>
		<button type="button" on:click={updateMeals}> Update Meal Status </button>
		<br />
		{#each localMeals as meal, index (meal.Meal)}
			<div>
				<input
					type="checkbox"
					id={meal.Meal}
					name={meal.Meal}
					value={meal.Meal}
					checked={meal.Enabled}
					on:change={() => toggleMeal(index)}
				/>
				<label for={meal.Meal}>
					{#if meal.URL}
						<a href={meal.URL} target="_blank" rel="noopener noreferrer">
							{meal.Meal}
						</a>
					{:else}
						{meal.Meal}
					{/if}
				</label>
			</div>
		{/each}
	</form>
{:else}
	<p>No meals found.</p>
{/if}
