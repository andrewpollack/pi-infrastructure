<script lang="ts">
	import type { MonthResponse } from '$lib/types';

	export let monthData: MonthResponse;
</script>

<h2>{monthData.Month} {monthData.Year}</h2>

<table border="1">
	<thead>
		<tr>
			<th>Sunday</th>
			<th>Monday</th>
			<th>Tuesday</th>
			<th>Wednesday</th>
			<th>Thursday</th>
			<th>Friday</th>
			<th>Saturday</th>
		</tr>
	</thead>
	<tbody>
		{#each monthData.MealsEachWeek as week}
			<tr>
				{#each week as meal}
					<td>
						{#if meal.URL}
							<a href={meal.URL} target="_blank" rel="noopener noreferrer">
								{meal.Meal}
							</a>
						{:else}
							{meal.Meal}
						{/if}
					</td>
				{/each}
			</tr>
		{/each}
	</tbody>
</table>
