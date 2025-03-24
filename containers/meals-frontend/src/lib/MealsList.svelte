<script lang="ts">
	import type { Meal } from '$lib/types';
	import { StatusType } from '$lib/types';
	import StatusIndicator from './StatusIndicator.svelte';

	let { meals }: { meals: Meal[] } = $props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);
	let localMeals = $state(meals.map((m) => ({ ...m })));
	let isDifferent = $derived(
		localMeals.some((meal, index) => meal.Enabled !== meals[index].Enabled)
	);

	function toggleMeal(index: number) {
		localMeals[index].Enabled = !localMeals[index].Enabled;
	}

	async function updateMeals() {
		message = 'Updating meal status...';
		statusType = StatusType.LOADING;

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
			message = 'Meal status updated!';
			statusType = StatusType.SUCCESS;
		} catch (error) {
			message = 'error updating meals: ';
			statusType = StatusType.ERROR;
			message += error instanceof Error ? error.message : String(error);
			alert(message);
		}
	}
</script>

<h2>Meal Status</h2>

<StatusIndicator {message} type={statusType} />

{#if localMeals && localMeals.length > 0}
	<form>
		<button type="button" disabled={!isDifferent} onclick={updateMeals}>
			Update Meal Status
		</button>
		<br />
		{#each localMeals as meal, index (meal.Meal)}
			<div>
				<input
					type="checkbox"
					id={meal.Meal}
					name={meal.Meal}
					value={meal.Meal}
					checked={meal.Enabled}
					onchange={() => toggleMeal(index)}
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
