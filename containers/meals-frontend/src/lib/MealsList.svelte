<script lang="ts">
    import type { Meal } from '$lib/types';
  
    export let meals: Meal[];
</script>

<h2>All Meals</h2>

{#if meals && meals.length > 0}
  <ul>
    {#each meals as meal}
      <li>
        {#if meal.URL}
          <a href={meal.URL} target="_blank" rel="noopener noreferrer">
            {meal.Meal}
          </a>
        {:else}
          {meal.Meal}
        {/if}
      </li>
    {/each}
  </ul>
{:else}
  <p>No meals found.</p>
{/if}
