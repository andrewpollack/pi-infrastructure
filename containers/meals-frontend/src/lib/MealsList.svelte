<script lang="ts">
	import type { Meal } from '$lib/types';

	export let meals: Meal[];
</script>

<h2>Disable Meals</h2>

{#if meals && meals.length > 0}
	<form>
		<input type="submit" value="Submit" /><br /><br />

		{#each meals as meal}
			<input type="checkbox" id={meal.Meal} name={meal.Meal} value={meal.Meal} />
			<label for={meal.Meal}>
				{#if meal.URL}
					<a href={meal.URL} target="_blank" rel="noopener noreferrer">
						{meal.Meal}
					</a>
				{:else}
					{meal.Meal}
				{/if}

				<br /></label
			>
		{/each}
	</form>
{:else}
	<p>No meals found.</p>
{/if}
