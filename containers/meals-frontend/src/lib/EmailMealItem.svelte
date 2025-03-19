<script lang="ts">
	import type { Meal } from '$lib/types';

	export let meal: Meal;
	export let isSelected: boolean;
	export let dayOfWeek: string;
	export let maxMeals: number;
	export let selectedMealsCount: number;
	export let onToggle: (meal: string, checked: boolean) => void;

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

<div class="meal-item">
	<label>
		<input type="checkbox" name="meals" checked={isSelected} on:change={handleChange} />
		{#if isSelected}
			(<strong>{dayOfWeek}</strong>)
		{/if}
		{#if meal.URL}
			<a href={meal.URL} target="_blank" rel="noopener noreferrer">
				{meal.Meal}
			</a>
		{:else}
			{meal.Meal}
		{/if}
	</label>
</div>
