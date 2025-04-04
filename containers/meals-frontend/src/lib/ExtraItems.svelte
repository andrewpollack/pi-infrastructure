<script lang="ts">
	import type { ExtraItem, ExtraItemUpdate } from '$lib/types';
	import { StatusType } from '$lib/types';
	import StatusIndicator from './StatusIndicator.svelte';
	import { Aisles, Color } from '$lib/const';
	import { onMount } from 'svelte';

	let { extraItems }: { extraItems: ExtraItem[] } = $props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);

	let localItems = $state(extraItems.map((m) => ({ ...m })));
	// This is a workaround to ensure that editingRows is initialized after localItems
	// is set. The onMount lifecycle function is called after the component is first
	// rendered, so we can safely use the value of localItems here.
	let editingRows = $state<boolean[]>([]);
	onMount(() => {
		editingRows = localItems.map(() => false);
	});

	let isDifferent = $derived(
		localItems.length !== extraItems.length ||
			localItems.some((item, index) => {
				const original = extraItems[index];
				return (
					!original ||
					original.ID !== item.ID ||
					original.Name !== item.Name ||
					original.Aisle !== item.Aisle
				);
			})
	);

	let hasEmptyName = $derived(localItems.some((item) => item.Name.trim().length === 0));
	let hasEmptyAisle = $derived(localItems.some((item) => item.Aisle.trim().length === 0));
	let isFormValid = $derived(!hasEmptyName && !hasEmptyAisle);
	let anyEditing = $derived(editingRows.some((isEd) => isEd));

	function isChanged(item: ExtraItem) {
		const original = extraItems.find((o) => o.ID === item.ID);
		if (!original) return true;
		return original.Name !== item.Name || original.Aisle !== item.Aisle;
	}

	function addItem() {
		// This is a temporary ID generation method. The DB will handle creating
		// unique IDs when the item is added to the database.
		// This is just to prevent overlap with existing IDs in the localItems array.
		localItems = [
			...localItems,
			{ ID: Math.floor(Math.random() * 10000) + 100, Name: '', Aisle: '' }
		];
		editingRows = [...editingRows, true];
	}

	function removeItem(index: number) {
		localItems = [...localItems.slice(0, index), ...localItems.slice(index + 1)];
		editingRows = [...editingRows.slice(0, index), ...editingRows.slice(index + 1)];
	}

	function toggleEditing(index: number) {
		editingRows = editingRows.map((val, i) => (i === index ? !val : val));
	}

	function buildChanges(): ExtraItemUpdate[] {
		const changes: ExtraItemUpdate[] = [];
		const oldMap = new Map<number, ExtraItem>();
		for (const oldItem of extraItems) {
			oldMap.set(oldItem.ID, oldItem);
		}
		const newMap = new Map<number, ExtraItem>();
		for (const newItem of localItems) {
			if (newItem.ID === 0) {
				changes.push({ Action: 'Add', Old: null, New: { ...newItem } });
			} else {
				newMap.set(newItem.ID, newItem);
			}
		}
		for (const oldItem of oldMap.values()) {
			const { ID } = oldItem;
			const updatedItem = newMap.get(ID);
			if (!updatedItem) {
				changes.push({ Action: 'Delete', Old: { ...oldItem }, New: null });
			} else {
				if (oldItem.Name !== updatedItem.Name || oldItem.Aisle !== updatedItem.Aisle) {
					changes.push({
						Action: 'Update',
						Old: { ...oldItem },
						New: { ...updatedItem }
					});
				}
				newMap.delete(ID);
			}
		}
		for (const item of newMap.values()) {
			changes.push({ Action: 'Add', Old: null, New: { ...item } });
		}
		return changes;
	}

	async function updateItems() {
		if (!isFormValid) {
			message = 'Error: One or more items have an empty name or aisle.';
			statusType = StatusType.ERROR;
			return;
		}
		if (anyEditing) {
			message = 'Please finish editing all items before updating.';
			statusType = StatusType.ERROR;
			return;
		}

		message = 'Updating items...';
		statusType = StatusType.LOADING;

		const changes = buildChanges();

		try {
			const res = await fetch('/api/items/update', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(changes)
			});
			await res.json();
			message = 'Items updated!';
			extraItems = [...localItems];
			statusType = StatusType.SUCCESS;
		} catch (error) {
			message = 'Error updating items: ' + (error instanceof Error ? error.message : String(error));
			statusType = StatusType.ERROR;
		}
	}
</script>

<StatusIndicator {message} type={statusType} />

<form class="reactive-font">
	<button
		type="button"
		onclick={updateItems}
		disabled={!isDifferent || anyEditing}
		class:warning={!isFormValid}
	>
		Update Items
	</button>
	<br /><br />
	<div
		class="table-responsive"
		style="--secondary-color: {Color.secondary}; --tertiary-color: {Color.tertiary}"
	>
		<table class="fixed-table">
			<colgroup>
				<col style="width: 2%; white-space: nowrap;" />
				<col style="width: 20%; white-space: nowrap;" />
				<col style="width: 20%; white-space: nowrap;" />
			</colgroup>
			<thead>
				<tr>
					<th>Edit/Delete</th>
					<th>Name</th>
					<th>Aisle</th>
				</tr>
			</thead>
			<tbody>
				{#each localItems as item, index}
					<tr>
						<td class="center-btn">
							<button
								class:unchanged={!isChanged(item)}
								type="button"
								onclick={() => toggleEditing(index)}
							>
								{editingRows[index] ? 'Done' : 'Edit'}
							</button>
							<button
								class="warning"
								style="opacity: 1.0;"
								type="button"
								onclick={() => removeItem(index)}
							>
								X
							</button>
						</td>

						<td class:unchanged={!isChanged(item)}>
							{#if editingRows[index]}
								<input class="cell-input" type="text" bind:value={item.Name} />
							{:else}
								{item.Name}
							{/if}
						</td>

						<td class:unchanged={!isChanged(item)} class="ellipsis">
							{#if editingRows[index]}
								<select class="cell-select" bind:value={item.Aisle}>
									{#each Aisles as aisle}
										<option value={aisle}>{aisle}</option>
									{/each}
								</select>
							{:else}
								{item.Aisle}
							{/if}
						</td>
					</tr>
				{/each}
				<tr>
					<td class="center-btn">
						<button type="button" onclick={addItem}> + </button>
					</td>
					<td></td>
					<td></td>
				</tr>
			</tbody>
		</table>
	</div>
</form>

<style>
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

	.unchanged {
		opacity: 0.6;
	}

	.warning {
		background-color: #e34234;
		color: white;
	}

	.cell-input,
	.cell-select {
		width: 100%;
		box-sizing: border-box;
	}

	.center-btn {
		text-align: center;
		white-space: nowrap; /* Keep buttons side-by-side */
	}

	.reactive-font {
		font-size: clamp(0.9rem, 2vw, 1.4rem);
	}

	.ellipsis {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	@media (max-width: 600px) {
		.fixed-table th,
		.fixed-table td {
			padding: 0.3rem;
			font-size: 0.9rem;
		}
	}

	th {
		background-color: var(--tertiary-color);
	}
</style>
