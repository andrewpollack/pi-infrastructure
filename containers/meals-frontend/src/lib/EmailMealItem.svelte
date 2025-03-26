<script lang="ts">
	import type { Meal } from '$lib/types';

	let {
		meal,
		isSelected,
		dayOfWeek,
		maxMeals,
		selectedMealsCount,
		onToggle,
		disableLinks
	}: {
		meal: Meal;
		isSelected: boolean;
		dayOfWeek: string;
		maxMeals: number;
		selectedMealsCount: number;
		onToggle: (meal: string, checked: boolean) => void;
		disableLinks: boolean;
	} = $props();

	function handleChange(e: Event) {
		const input = e.target as HTMLInputElement;
		if (input.checked && selectedMealsCount >= maxMeals) {
			input.checked = false;
			alert(`error: You can only select up to ${maxMeals} meals.`);
			return;
		}
		onToggle(meal.Meal, input.checked);
	}
</script>

<div>
	<label style="display: flex; align-items: center;">
		<input
			type="checkbox"
			name="meals"
			checked={isSelected}
			onchange={handleChange}
			style="margin-right: 15px;"
		/>
		{#if meal.URL && !disableLinks}
			<a href={meal.URL} target="_blank" rel="noopener noreferrer">
				{#if isSelected}
					<span>(<strong>{dayOfWeek}</strong>)</span>
				{/if}
				{meal.Meal}
			</a>
		{:else}
			<span>
				{#if isSelected}
					<span>(<strong>{dayOfWeek}</strong>)</span>
				{/if}
				{meal.Meal}</span
			>
		{/if}
	</label>
</div>
