<script lang="ts">
	import type { Meal } from '$lib/types';
	import { StatusType } from '$lib/types';
	import StatusIndicator from './StatusIndicator.svelte';
	import { Color } from '$lib/const';

	let { meals }: { meals: Meal[] } = $props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);

	let localMeals = $state(meals.map((m) => ({ ...m })));

	function isMealChanged(meal: Meal, index: number) {
		return meal.Enabled !== meals[index].Enabled;
	}

	let isDifferent = $derived(localMeals.some((meal, index) => isMealChanged(meal, index)));

	async function updateMeals() {
		message = 'Updating meal status...';
		statusType = StatusType.LOADING;

		const updates = localMeals
			.filter((meal, index) => isMealChanged(meal, index))
			.map((m) => ({ name: m.Meal, disabled: !m.Enabled }));

		try {
			const res = await fetch('/api/enable', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(updates)
			});
			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending data.');
			}
			message = 'Meal status updated!';
			statusType = StatusType.SUCCESS;

			meals = [...localMeals];
		} catch (error) {
			message = 'Error updating meals: ' + (error instanceof Error ? error.message : String(error));
			statusType = StatusType.ERROR;
			alert(message);
		}
	}
</script>

<StatusIndicator {message} type={statusType} />

{#if localMeals && localMeals.length > 0}
	<form class="reactive-font">
		<button type="button" onclick={updateMeals} disabled={!isDifferent}>
			Update Meal Status
		</button>

		<br /><br />

		<div
			class="table-responsive"
			style="--secondary-color: {Color.secondary}; --tertiary-color: {Color.tertiary}"
		>
			<table class="fixed-table">
				<colgroup>
					<col style="width: 2%; white-space: nowrap;" />
					<col style="width: 90%; white-space: nowrap;" />
				</colgroup>
				<thead>
					<tr>
						<th>Enabled</th>
						<th>Meal</th>
					</tr>
				</thead>
				<tbody>
					{#each localMeals as meal, index (meal.Meal)}
						<tr class:unchanged={!isMealChanged(meal, index)}>
							<td class="center-btn">
								<input
									type="checkbox"
									id={meal.Meal}
									name={meal.Meal}
									bind:checked={meal.Enabled}
								/>
							</td>

							<td>
								<label for={meal.Meal}>
									{#if meal.URL}
										<a href={meal.URL} target="_blank" rel="noopener noreferrer">
											{meal.Meal}
										</a>
									{:else}
										{meal.Meal}
									{/if}
								</label>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</form>
{:else}
	<p>No meals found.</p>
{/if}

<style>
	.reactive-font {
		font-size: clamp(0.9rem, 1vw, 1.2rem);
	}

	.table-responsive {
		width: 100%;
		overflow-x: auto;
	}

	.fixed-table {
		width: 100%;
		border-collapse: collapse;
	}

	.fixed-table th,
	.fixed-table td {
		padding: 0.5rem;
		border: 1px solid var(--secondary-color);
	}

	th {
		background-color: var(--tertiary-color);
	}

	.center-btn {
		text-align: center;
		white-space: nowrap;
	}

	.unchanged {
		opacity: 0.6;
	}

	@media (max-width: 600px) {
		.fixed-table th,
		.fixed-table td {
			padding: 0.3rem;
			font-size: 0.9rem;
		}
	}
</style>
