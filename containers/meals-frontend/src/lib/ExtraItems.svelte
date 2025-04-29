<script lang="ts">
	import type { ExtraItem, ExtraItemUpdate } from '$lib/types';
	import { StatusType } from '$lib/types';
	import StatusIndicator from './StatusIndicator.svelte';
	import { Aisles, Color } from '$lib/const';

	let { extraItems }: { extraItems: ExtraItem[] } = $props();

	let message = $state('');
	let statusType = $state(StatusType.SUCCESS);

	// "localItems" holds all items in the UI
	let localItems = $state(extraItems.map((m) => ({ ...m })));

	// A single toggle to control edit mode for ALL rows
	let isEditingAll = $state(false);

	function itemsDifferent(item: ExtraItem, other: ExtraItem) {
		if (!item || !other) return true;
		return item.Name !== other.Name || item.Aisle !== other.Aisle || item.Enabled !== other.Enabled;
	}

	let isDifferent = $derived(
		localItems.length !== extraItems.length ||
			localItems.some((item, index) => {
				const original = extraItems[index];
				return itemsDifferent(item, original);
			})
	);

	// Basic validation
	let hasEmptyName = $derived(localItems.some((item) => item.Name.trim().length === 0));
	let hasEmptyAisle = $derived(localItems.some((item) => item.Aisle.trim().length === 0));
	let isFormValid = $derived(!hasEmptyName && !hasEmptyAisle);

	function isChanged(item: ExtraItem) {
		const original = extraItems.find((o) => o.ID === item.ID);
		if (!original) return true;
		return (
			original.Name !== item.Name ||
			original.Aisle !== item.Aisle ||
			original.Enabled !== item.Enabled
		);
	}

	function toggleEditAll() {
		isEditingAll = !isEditingAll;
	}

	function addItem() {
		// Force edit mode on if it's not already
		if (!isEditingAll) {
			isEditingAll = true;
		}

		// Temporary ID generation. The DB will handle unique IDs when adding.
		localItems = [
			...localItems,
			{ ID: Math.floor(Math.random() * 10000) + 100, Name: '', Aisle: '', Enabled: true }
		];
	}

	function removeItem(index: number) {
		localItems = [...localItems.slice(0, index), ...localItems.slice(index + 1)];
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

		// Identify deleted or updated items
		for (const oldItem of oldMap.values()) {
			const { ID } = oldItem;
			const updatedItem = newMap.get(ID);
			if (!updatedItem) {
				changes.push({ Action: 'Delete', Old: { ...oldItem }, New: null });
			} else {
				if (itemsDifferent(oldItem, updatedItem)) {
					changes.push({
						Action: 'Update',
						Old: { ...oldItem },
						New: { ...updatedItem }
					});
				}
				newMap.delete(ID);
			}
		}
		// Any remaining items in newMap are genuinely new
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

<!-- Toggle all rows editable or read-only -->
<button type="button" onclick={toggleEditAll}>
	{isEditingAll ? 'Done Editing' : 'Edit Items'}
</button>

<button type="button" onclick={updateItems} disabled={!isDifferent} class:warning={!isFormValid}>
	Update Items
</button>

<br /><br />
<div
	class="table-responsive"
	style="--secondary-color: {Color.secondary}; --tertiary-color: {Color.tertiary}"
>
	<table class="fixed-table">
		<colgroup>
			<col style="width: 2%;" />
			<col style="width: 20%;" />
			<col style="width: 20%;" />
		</colgroup>
		<thead>
			<tr>
				<th>Enabled / Delete</th>
				<th>Name</th>
				<th>Aisle</th>
			</tr>
		</thead>
		<tbody>
			{#each localItems as item, index}
				<tr>
					<td class="center-btn">
						<!-- Enabled checkbox is only editable if isEditingAll is true -->
						{#if isEditingAll}
							<input type="checkbox" bind:checked={item.Enabled} />
						{:else}
							<input type="checkbox" checked={item.Enabled} disabled />
						{/if}

						<!-- Delete button, only active if editingAll is true -->
						<button
							class="warning"
							type="button"
							onclick={() => removeItem(index)}
							disabled={!isEditingAll}
						>
							X
						</button>
					</td>

					<td class:unchanged={!isChanged(item)}>
						{#if isEditingAll}
							<input class="cell-input" type="text" bind:value={item.Name} />
						{:else}
							{item.Name}
						{/if}
					</td>

					<td class:unchanged={!isChanged(item)} class="ellipsis">
						{#if isEditingAll}
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

			<!-- Always show "Add Item," but it automatically switches to edit mode. -->
			<tr>
				<td colspan="3" class="center-btn">
					<button type="button" onclick={addItem}> + </button>
				</td>
			</tr>
		</tbody>
	</table>
</div>

<style>
	.table-responsive {
		width: 100%;
		overflow-x: auto;
		font-size: clamp(0.9rem, 1vw, 1.4rem);
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

	.unchanged {
		opacity: 0.6;
	}

	.warning {
		background-color: #e34234;
		color: white;
	}

	.warning:disabled {
		background-color: #999 !important;
		color: #fff !important;
		cursor: not-allowed;
		opacity: 0.8; /* tweak as desired */
	}

	.cell-input,
	.cell-select {
		width: 100%;
		box-sizing: border-box;
	}

	.center-btn {
		text-align: center;
		white-space: nowrap;
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
</style>
