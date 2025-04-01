<script lang="ts">
	import type { Meal, ExtraItem } from '$lib/types';
	import { StatusType } from '$lib/types';
	import { DaysOfWeek, Color } from '$lib/const';
	import EmailMealItem from './EmailMealItem.svelte';
	import StatusIndicator from './StatusIndicator.svelte';

	let { meals, emails, extraItems }: { meals: Meal[]; emails: string[]; extraItems: ExtraItem[] } =
		$props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);
	let selectedMeals = $state([] as string[]);
	let selectedEmails = $state([] as string[]);
	let selectedExtraItems = $state([] as string[]);
	let disableLinks: boolean = $state(false);

	let isEmailSelected = $derived(selectedEmails.length > 0);
	let isMealsSelected = $derived(selectedMeals.length > 0);

	const maxMeals = 7;

	function toggleMeal(meal: string, checked: boolean) {
		if (checked) {
			selectedMeals = [...selectedMeals, meal];
		} else {
			selectedMeals = selectedMeals.filter((m) => m !== meal);
		}
	}

	function getChunkedArray<T>(array: T[], chunkSize: number): T[][] {
		const result: T[][] = [];
		for (let i = 0; i < array.length; i += chunkSize) {
			result.push(array.slice(i, i + chunkSize));
		}
		return result;
	}

	const allMealItems: Meal[] = [
		...[
			{ Day: 0, Meal: 'Out', URL: null, Enabled: null },
			{ Day: 0, Meal: 'Leftovers', URL: null, Enabled: null }
		],
		...meals
	];
	const numColumns = 2;
	const chunkedMeals: Meal[][] = getChunkedArray(
		allMealItems,
		Math.ceil(allMealItems.length / numColumns)
	);
	const chunkedExtraItems: ExtraItem[][] = getChunkedArray(
		extraItems,
		Math.ceil(extraItems.length / numColumns)
	);

	function getDayLabel(mealName: string) {
		const index = selectedMeals.indexOf(mealName);
		return index > -1 ? DaysOfWeek[index] : '';
	}

	function handleMealChange(meal: Meal, event: Event) {
		const input = event.target as HTMLInputElement;
		if (input.checked && selectedMeals.length >= maxMeals) {
			input.checked = false;
			alert(`Error: You can only select up to ${maxMeals} meals.`);
			return;
		}
		toggleMeal(meal.Meal, input.checked);
	}

	async function handleSubmit(event: Event) {
		event.preventDefault();
		message = 'Sending email...';
		statusType = StatusType.LOADING;
		let paddedMeals = selectedMeals;

		if (paddedMeals.length < maxMeals) {
			paddedMeals = [...paddedMeals, ...Array(maxMeals - paddedMeals.length).fill('Out')];
		}

		try {
			const res = await fetch('/api/email', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					meals: paddedMeals,
					emails: selectedEmails,
					extraItems: selectedExtraItems
				})
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending data.');
			}

			await res.json();
			message = 'Email sent successfully!';
			statusType = StatusType.SUCCESS;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			message = `Error sending email: ${errorMessage}`;
			statusType = StatusType.ERROR;
		}
	}
</script>

<h2>Email</h2>

<StatusIndicator {message} type={statusType} />

<div
	class="table-container"
	style="--primary-color: {Color.primary}; --secondary-color: {Color.secondary}; --tertiary-color: {Color.tertiary}"
>
	<form onsubmit={handleSubmit}>
		<button disabled={!isEmailSelected || !isMealsSelected} type="submit"> Send Email </button>

		<br /><br />

		<table>
			<thead>
				<tr class="header-row">
					{#each DaysOfWeek as day}
						<th>{day}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				<tr>
					{#each Array(maxMeals) as _, i}
						<td class="meal-cell">
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

		<!-- Group Emails using fieldset/legend for better semantics -->
		<fieldset>
			<legend><strong> Emails </strong></legend>
			{#each emails as email}
				<div>
					<label class="checkbox-label">
						<input type="checkbox" value={email} bind:group={selectedEmails} />
						{email}
					</label>
				</div>
			{/each}
		</fieldset>

		<!-- Group Extra Items -->
		<fieldset>
			<legend><strong> Extra Items </strong></legend>
			<div class="extra-items">
				{#each chunkedExtraItems as chunk}
					<div class="extra-column">
						{#each chunk as item}
							<div>
								<label class="checkbox-label">
									<input type="checkbox" value={item.Name} bind:group={selectedExtraItems} />
									{item.Name}
								</label>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		</fieldset>

		<!-- Group Meals -->
		<fieldset>
			<legend>
				<div>
					<strong> Meals </strong>
					<input type="checkbox" bind:checked={disableLinks} />
					Disable Hyperlinks
				</div>
			</legend>
			<div class="meals-list">
				{#each chunkedMeals as chunk}
					<div class="meal-column">
						{#each chunk as meal}
							<div>
								<label class="checkbox-label">
									<input
										type="checkbox"
										name="meals"
										checked={selectedMeals.includes(meal.Meal)}
										onchange={(e) => handleMealChange(meal, e)}
									/>
									{#if meal.URL && !disableLinks}
										<a href={meal.URL} target="_blank" rel="noopener noreferrer">
											{#if selectedMeals.includes(meal.Meal)}
												<span>(<strong>{getDayLabel(meal.Meal)}</strong>)</span>
											{/if}
											{meal.Meal}
										</a>
									{:else}
										<span>
											{#if selectedMeals.includes(meal.Meal)}
												<span>(<strong>{getDayLabel(meal.Meal)}</strong>)</span>
											{/if}
											{meal.Meal}
										</span>
									{/if}
								</label>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		</fieldset>
	</form>
</div>

<style>
	.table-container {
		max-width: 100%;
	}

	table {
		width: 100%;
		border-collapse: collapse;
		table-layout: fixed;
	}

	table,
	th,
	td {
		border: 1px solid var(--secondary-color);
	}

	.header-row {
		background-color: var(--tertiary-color);
	}

	th,
	td {
		padding: 0.5rem;
		text-align: center;
		vertical-align: top;
		font-size: clamp(0.7rem, 2vw, 1rem);
		word-wrap: break-word;
	}

	.meal-cell {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: 15px;
	}

	.extra-items,
	.meals-list {
		display: flex;
		gap: 1rem;
		flex-wrap: nowrap;
	}

	.extra-column,
	.meal-column {
		flex: 0 0 50%;
	}

	fieldset {
		border: 1px solid var(--secondary-color);
	}
</style>
