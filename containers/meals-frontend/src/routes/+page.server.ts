import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { CalendarResponse, MealsResponse } from '$lib/types';

export const load: PageServerLoad = async () => {
	const emails = env.EMAILS
		? env.EMAILS.split(',').map((email) => email.trim())
		: ['user@example.com'];

	try {
		const [mealsRes, calendarRes] = await Promise.all([
			fetch(`${env.API_BASE_URL}/api/meals`),
			fetch(`${env.API_BASE_URL}/api/calendar`)
		]);

		// Check for non-OK responses
		if (!mealsRes.ok) {
			throw error(mealsRes.status, 'Failed to fetch meals');
		}
		if (!calendarRes.ok) {
			throw error(calendarRes.status, 'Failed to fetch calendar');
		}

		const mealsData: MealsResponse = await mealsRes.json();
		const calendarData: CalendarResponse = await calendarRes.json();

		return {
			allMeals: mealsData.allMeals,
			currMonthResponse: calendarData.currMonthResponse,
			allEmails: emails
		};
	} catch (err: any) {
		console.error('Error loading page data:', err);
		throw error(500, 'Unable to load page data at the moment. Please try again later.');
	}
};
