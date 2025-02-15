import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { MealsResponse, CalendarResponse } from '$lib/types';
import fs from 'fs/promises';

export const load: PageServerLoad = async () => {
	let mealsData: MealsResponse;
	let allCalendars: CalendarResponse;

	const isDev = !env.API_BASE_URL;

	if (isDev) {
		const mealsFile = await fs.readFile('src/lib/data/meals.json', 'utf-8');
		const calendarFile = await fs.readFile('src/lib/data/calendar.json', 'utf-8');

		mealsData = JSON.parse(mealsFile);
		allCalendars = JSON.parse(calendarFile);
	} else {
		const mealsRes = await fetch(`${env.API_BASE_URL}/api/meals`);
		mealsData = await mealsRes.json();

		const calendarRes = await fetch(`${env.API_BASE_URL}/api/calendar`);
		allCalendars = await calendarRes.json();
	}

	return {
		isDev,
		allMeals: mealsData.allMeals,
		allCalendars
	};
};
