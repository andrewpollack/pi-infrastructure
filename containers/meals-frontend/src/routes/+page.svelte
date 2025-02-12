<script lang="ts">
  import type { PageData } from './$types';

  // Import our new components
  import CalendarMonth from '$lib/CalendarMonth.svelte';
  import MealsList from '$lib/MealsList.svelte';

  export let data: PageData;

  // Destructure props for convenience
  const { allMeals, allCalendars } = data;
</script>

<h1>Meals</h1>

<!-- Render the "Current" month -->
<CalendarMonth 
  monthData={allCalendars.currMonthResponse} 
/>

<!-- Render the "Next" month -->
<CalendarMonth 
  monthData={allCalendars.nextMonthResponse} 
/>

<!-- Render the "All Meals" list -->
<MealsList meals={allMeals} />
