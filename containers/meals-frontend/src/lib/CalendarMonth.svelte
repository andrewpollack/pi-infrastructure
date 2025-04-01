<script lang="ts">
	import type { MonthResponse } from '$lib/types';
	import { DaysOfWeek, Color } from '$lib/const';

	let { monthData }: { monthData: MonthResponse } = $props();
</script>

<h2>{monthData.Month} {monthData.Year}</h2>

<div
	class="table-container"
	style="--primary-color: {Color.primary};
		   --secondary-color: {Color.secondary};
		   --tertiary-color: {Color.tertiary};
		   --quaternary-color: {Color.quaternary};"
>
	<table>
		<thead>
			<tr>
				{#each DaysOfWeek as day}
					<th>{day}</th>
				{/each}
			</tr>
		</thead>
		<tbody>
			{#each monthData.MealsEachWeek as week}
				<tr>
					{#each week as meal}
						<td>
							{#if meal.Day > 0}
								<div class="number-header">
									<strong>{meal.Day}</strong>
								</div>
							{/if}

							<div class="meal-content">
								{#if meal.URL}
									<a href={meal.URL} target="_blank" rel="noopener noreferrer">
										{meal.Meal}
									</a>
								{:else}
									{meal.Meal}
								{/if}
							</div>
						</td>
					{/each}
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.table-container {
		max-width: 100%;
	}

	table {
		width: 100%;
		table-layout: fixed;
		border-collapse: collapse;
	}

	thead {
		background-color: var(--tertiary-color);
	}

	th,
	td {
		border: 1px solid var(--secondary-color);
		padding: 0;
		text-align: center;
		vertical-align: top;
		word-wrap: break-word;
		white-space: normal;

		/*
		 * clamp(min, preferred, max)
		 *  - min font size: 0.7rem
		 *  - let the browser choose in between based on available space (2vw)
		 *  - max font size: 1rem
		 */
		font-size: clamp(0.7rem, 2vw, 1rem);
	}

	tbody td {
		height: 60px;
		vertical-align: top;
	}

	.number-header {
		display: block;
		width: 100%;
		background-color: var(--quaternary-color);
		font-size: clamp(0.8rem, 2vw, 1.1rem);
		border-bottom: 1px solid #ccc;
	}

	.meal-content {
		display: -webkit-box;
		-webkit-box-orient: vertical;
		line-clamp: 2;
		overflow: hidden;
		text-overflow: ellipsis;
		padding: 0.2rem;
		height: calc(100% - 1em);
		box-sizing: border-box;
	}
</style>
