import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { MealsResponse, CalendarResponse } from '$lib/types';

export const load: PageServerLoad = async () => {
  // 1) Fetch data from /api/meals
  const API_BASE_URL = env.API_BASE_URL;

  const mealsRes = await fetch(`${API_BASE_URL}/api/meals`);
  const mealsData: MealsResponse = await mealsRes.json();

  // 2) Fetch data from /api/calendar
  const calendarRes = await fetch(`${API_BASE_URL}/api/calendar`);
  const calendarData: CalendarResponse = await calendarRes.json();

  // Return the combined data to the Svelte page
  return {
    allMeals: mealsData.allMeals,
    allCalendars: calendarData
  };
};
