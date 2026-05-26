<script lang="ts">
	import type { Meal, ExtraItem, AisleGroup, OneOffItem } from '$lib/types';
	import { StatusType } from '$lib/types';
	import { DaysOfWeek, Color } from '$lib/const';
	import StatusIndicator from './StatusIndicator.svelte';
	import { emailState } from '$lib/emailState.svelte';

	let {
		meals,
		emails,
		extraItems,
		aisles
	}: { meals: Meal[]; emails: string[]; extraItems: ExtraItem[]; aisles: string[] } = $props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);

	let selectedMeals = $derived.by(() => emailState.selectedMeals);
	let selectedEmails = $derived.by(() => emailState.selectedEmails);
	let selectedExtraItems = $derived.by(() => emailState.selectedExtraItems);
	let disableLinks = $derived.by(() => emailState.disableLinks);

	// Step 1 = meal/email selection, Step 2 = ingredient preview
	let step = $state(1);
	let aisleGroups = $state([] as AisleGroup[]);
	let excludedItems = $state(new Set<string>());

	// One-off items: keyed by aisle name → list of item names
	let oneOffItems = $state<Record<string, string[]>>({});
	// Excluded one-off items: keyed by "aisle::name"
	let excludedOneOffItems = $state(new Set<string>());
	// Which aisle's inline add-form is open
	let addingToAisle = $state<string | null>(null);
	let newItemText = $state('');

	// Remove any persisted extra-item selections that no longer exist in the DB.
	$effect(() => {
		const valid = new Set(extraItems.map((i) => i.Name));
		const cleaned = emailState.selectedExtraItems.filter((n) => valid.has(n));
		if (cleaned.length !== emailState.selectedExtraItems.length) {
			emailState.selectedExtraItems = cleaned;
		}
	});

	let isEmailSelected = $derived(selectedEmails.length > 0);
	let isMealsSelected = $derived(selectedMeals.length > 0);

	const maxMeals = 7;

	function toggleMeal(meal: string, checked: boolean) {
		if (checked) {
			emailState.selectedMeals = [...emailState.selectedMeals, meal];
		} else {
			emailState.selectedMeals = emailState.selectedMeals.filter((m) => m !== meal);
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
		if (input.checked && emailState.selectedMeals.length >= maxMeals) {
			input.checked = false;
			alert(`Error: You can only select up to ${maxMeals} meals.`);
			return;
		}
		toggleMeal(meal.Meal, input.checked);
	}

	function getPaddedMeals(): string[] {
		const padded = [...selectedMeals];
		if (padded.length < maxMeals) {
			padded.push(...Array(maxMeals - padded.length).fill('Out'));
		}
		return padded;
	}

	async function handlePreview(event: Event) {
		event.preventDefault();
		message = 'Loading grocery list...';
		statusType = StatusType.LOADING;

		try {
			const res = await fetch('/api/email/preview', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					meals: getPaddedMeals(),
					extraItems: selectedExtraItems
				})
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while loading preview.');
			}

			const data = await res.json();
			aisleGroups = data.aisleGroups ?? [];
			excludedItems = new Set<string>();
			message = '';
			step = 2;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			message = `Error loading preview: ${errorMessage}`;
			statusType = StatusType.ERROR;
		}
	}

	function toggleItem(itemName: string, checked: boolean) {
		const next = new Set(excludedItems);
		if (!checked) {
			next.add(itemName);
		} else {
			next.delete(itemName);
		}
		excludedItems = next;
	}

	function startAddingToAisle(aisle: string) {
		addingToAisle = aisle;
		newItemText = '';
	}

	function cancelAddItem() {
		addingToAisle = null;
		newItemText = '';
	}

	function confirmAddItem(aisle: string) {
		const name = newItemText.trim();
		if (!name) return;
		oneOffItems[aisle] = [...(oneOffItems[aisle] ?? []), name];
		addingToAisle = null;
		newItemText = '';
	}

	function toggleOneOffItem(aisle: string, name: string, checked: boolean) {
		const key = `${aisle}::${name}`;
		const next = new Set(excludedOneOffItems);
		if (!checked) {
			next.add(key);
		} else {
			next.delete(key);
		}
		excludedOneOffItems = next;
	}

	function getOneOffItemsForSend(): OneOffItem[] {
		return Object.entries(oneOffItems).flatMap(([aisle, names]) =>
			names
				.filter((name) => !excludedOneOffItems.has(`${aisle}::${name}`))
				.map((name) => ({ name, aisle }))
		);
	}

	async function handleSend(event: Event) {
		event.preventDefault();
		message = 'Sending email...';
		statusType = StatusType.LOADING;

		try {
			const res = await fetch('/api/email', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					meals: getPaddedMeals(),
					emails: selectedEmails,
					extraItems: selectedExtraItems,
					excludedItems: Array.from(excludedItems),
					oneOffItems: getOneOffItemsForSend()
				})
			});

			if (!res.ok) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while sending data.');
			}

			await res.json();
			message = 'Email sent successfully!';
			statusType = StatusType.SUCCESS;
			step = 1;
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
	{#if step === 1}
		<form onsubmit={handlePreview}>
			<button disabled={!isEmailSelected || !isMealsSelected} type="submit">
				Review Grocery List
			</button>

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
						{#each [...Array(maxMeals).keys()] as i}
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
							<input type="checkbox" value={email} bind:group={emailState.selectedEmails} />
							{email}
						</label>
					</div>
				{/each}
			</fieldset>

			<!-- Group Extra Items -->
			{#if extraItems.length > 0}
				<fieldset>
					<legend><strong> Extra Items </strong></legend>
					<div class="extra-items">
						{#each chunkedExtraItems as chunk}
							<div class="extra-column">
								{#each chunk as item}
									<div>
										<label class="checkbox-label">
											<input
												type="checkbox"
												value={item.Name}
												bind:group={emailState.selectedExtraItems}
											/>
											{item.Name}
										</label>
									</div>
								{/each}
							</div>
						{/each}
					</div>
				</fieldset>
			{/if}

			<!-- Group Meals -->
			<fieldset>
				<legend>
					<div>
						<strong> Meals </strong>
						<label>
							<input type="checkbox" bind:checked={emailState.disableLinks} />
							Disable Links
						</label>
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
	{:else}
		<form onsubmit={handleSend}>
			<div class="preview-actions">
				<button
					type="button"
					onclick={() => {
						step = 1;
						message = '';
						oneOffItems = {};
						excludedOneOffItems = new Set();
						addingToAisle = null;
						newItemText = '';
					}}
				>
					Back
				</button>
				<button type="submit">Send Email</button>
			</div>

			<br />

			{#each aisles as aisle}
				{@const aisleGroup = aisleGroups.find((g) => g.aisle === aisle) ?? null}
				{@const oneOffs = oneOffItems[aisle] ?? []}
				<fieldset>
					<legend><strong>{aisle}</strong></legend>

					{#if aisleGroup}
						{#each aisleGroup.ingredients as ingredient}
							{@const colonIdx = ingredient.display.indexOf(': ')}
							{@const hasQty = colonIdx !== -1}
							{@const qtyUnit = hasQty ? ingredient.display.slice(0, colonIdx) : ''}
							{@const rest = hasQty ? ingredient.display.slice(colonIdx + 2) : ingredient.display}
							{@const parenIdx = rest.lastIndexOf(' (')}
							{@const itemName = parenIdx !== -1 ? rest.slice(0, parenIdx) : rest}
							{@const itemMeals = parenIdx !== -1 ? rest.slice(parenIdx + 1) : ''}
							<div>
								<label class="checkbox-label">
									<input
										type="checkbox"
										checked={!excludedItems.has(ingredient.name)}
										onchange={(e) =>
											toggleItem(ingredient.name, (e.target as HTMLInputElement).checked)}
									/>
									<span class="ingredient-label">
										{#if hasQty}
											<strong class="qty">{qtyUnit}</strong>
										{/if}
										<span class="item-name">{itemName}</span>
										{#if itemMeals}
											<span class="item-meals">{itemMeals}</span>
										{/if}
									</span>
								</label>
							</div>
						{/each}
					{/if}

					{#each oneOffs as oneOff}
						<div>
							<label class="checkbox-label">
								<input
									type="checkbox"
									checked={!excludedOneOffItems.has(`${aisle}::${oneOff}`)}
									onchange={(e) =>
										toggleOneOffItem(aisle, oneOff, (e.target as HTMLInputElement).checked)}
								/>
								<span class="ingredient-label">
									<span class="item-name">{oneOff}</span>
									<span class="item-meals one-off-tag">(one-off)</span>
								</span>
							</label>
						</div>
					{/each}

					{#if addingToAisle === aisle}
						<div class="add-oneoff-form">
							<input
								type="text"
								bind:value={newItemText}
								placeholder="Item name..."
								autofocus
								onkeydown={(e) => {
									if (e.key === 'Enter') {
										e.preventDefault();
										confirmAddItem(aisle);
									} else if (e.key === 'Escape') {
										cancelAddItem();
									}
								}}
							/>
							<button type="button" onclick={() => confirmAddItem(aisle)}>Add</button>
							<button type="button" onclick={cancelAddItem}>Cancel</button>
						</div>
					{:else}
						<button type="button" class="add-oneoff-btn" onclick={() => startAddingToAisle(aisle)}
							>+</button
						>
					{/if}
				</fieldset>
			{/each}
		</form>
	{/if}
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
		text-align: left;
	}

	.preview-actions {
		display: flex;
		gap: 1rem;
	}

	.ingredient-label {
		display: flex;
		align-items: baseline;
		gap: 0.4rem;
		flex-wrap: wrap;
	}

	.qty {
		min-width: 4rem;
		display: inline-block;
	}

	.item-name {
		flex: 1;
	}

	.item-meals {
		font-size: 0.8em;
		color: #888;
	}

	.one-off-tag {
		font-size: 0.75em;
		color: #aaa;
		font-style: italic;
	}

	.add-oneoff-btn {
		display: block;
		margin-top: 0.4rem;
		padding: 0.1rem 0.6rem;
		font-size: 1rem;
		cursor: pointer;
		border: 1px solid var(--secondary-color);
		border-radius: 3px;
		background: transparent;
	}

	.add-oneoff-btn:hover {
		background-color: var(--tertiary-color);
	}

	.add-oneoff-form {
		display: flex;
		align-items: center;
		gap: 0.4rem;
		margin-top: 0.4rem;
	}

	.add-oneoff-form input[type='text'] {
		flex: 1;
		padding: 0.2rem 0.4rem;
		font-size: 0.9rem;
		border: 1px solid var(--secondary-color);
		border-radius: 3px;
	}
</style>
