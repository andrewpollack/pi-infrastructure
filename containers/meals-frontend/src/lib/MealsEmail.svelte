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

	async function handleSubmit(event: Event) {
		event.preventDefault();
		message = 'Sending email...';
		statusType = StatusType.LOADING;
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
					emails: selectedEmails,
					extraItems: selectedExtraItems
				})
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending data.');
			}

			const data = await res.json();
			message = 'Email sent successfully!';
			statusType = StatusType.SUCCESS;
		} catch (error) {
			message = 'error sending email: ' + (error instanceof Error ? error.message : String(error));
			statusType = StatusType.ERROR;
			if (error instanceof Error) {
				message += error.message;
			} else {
				message += String(error);
			}
			alert(message);
		}
	}
</script>

<h2>Email</h2>

<StatusIndicator {message} type={statusType} />

<div class="table-container">
	<form onsubmit={handleSubmit}>
		<button disabled={!isEmailSelected || !isMealsSelected} type="submit">Send Email</button>

		<br /> <br />

		<table border="1" style="border-collapse: collapse;">
			<thead>
				<tr style="background-color: {Color.tertiary};">
					{#each DaysOfWeek as day}
						<th>{day}</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				<tr>
					{#each [...Array(maxMeals).keys()] as i}
						<td
							style="
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
					<label style="display: flex; align-items: center;">
						<input
							type="checkbox"
							value={email}
							bind:group={selectedEmails}
							style="margin-right: 15px;"
						/>
						{email}
					</label>
				</div>
			{/each}
		</div>

		<div>
			<h3>Extra Items</h3>
			<div style="display: flex; gap: 1rem; flex-wrap: nowrap;">
				{#each chunkedExtraItems as chunk}
					<div style="flex: 0 0 50%;">
						{#each chunk as item}
							<div>
								<label style="display: flex; align-items: center;">
									<input
										type="checkbox"
										value={item.Name}
										bind:group={selectedExtraItems}
										style="margin-right: 15px;"
									/>
									{item.Name}
								</label>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		</div>

		<div>
			<div style="display: flex; align-items: center; gap: 1rem;">
				<h3>Meals</h3>
				<label>
					<input type="checkbox" bind:checked={disableLinks} />
					Disable Hyperlinks
				</label>
			</div>

			<div style="display: flex; gap: 1rem; flex-wrap: nowrap;">
				{#each chunkedMeals as chunk}
					<div style="flex: 0 0 50%;">
						{#each chunk as meal}
							<EmailMealItem
								{meal}
								isSelected={selectedMeals.includes(meal.Meal)}
								dayOfWeek={DaysOfWeek[selectedMeals.indexOf(meal.Meal)]}
								{maxMeals}
								selectedMealsCount={selectedMeals.length}
								onToggle={toggleMeal}
								{disableLinks}
							/>
						{/each}
					</div>
				{/each}
			</div>
		</div>
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

	th,
	td {
		word-wrap: break-word;
		white-space: normal;
		text-align: center;
		vertical-align: top;
		padding: 0.5rem;

		/* 
		 * clamp(min, preferred, max)
		 *  - min font size: 0.70rem
		 *  - let the browser choose in between based on available space via 2vw
		 *  - max font size: 1rem
		 */
		font-size: clamp(0.7rem, 2vw, 1rem);
	}
</style>
