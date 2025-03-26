import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { CalendarResponse, MealsResponse } from '$lib/types';

export const load: PageServerLoad = async () => {
	const emails = env.EMAILS
		? env.EMAILS.split(',').map((email) => email.trim())
		: ['user@example.com'];

	try {
		const [mealsRes, calendarRes, extraItemsRes] = await Promise.all([
			fetch(`${env.API_BASE_URL}/api/meals`),
			fetch(`${env.API_BASE_URL}/api/calendar`),
			fetch(`${env.API_BASE_URL}/api/items`)
		]);

		if (!mealsRes.ok) {
			throw error(mealsRes.status, 'Failed to fetch meals');
		}
		if (!calendarRes.ok) {
			throw error(calendarRes.status, 'Failed to fetch calendar');
		}
		if (!extraItemsRes.ok) {
			throw error(extraItemsRes.status, 'Failed to fetch extra items');
		}

		const mealsData: MealsResponse = await mealsRes.json();
		const calendarData: CalendarResponse = await calendarRes.json();
		const extraItemsData = await extraItemsRes.json();

		return {
			allMeals: mealsData.allMeals,
			currMonthResponse: calendarData.currMonthResponse,
			allEmails: emails,
			allExtraItems: extraItemsData.allItems
		};
	} catch (err: any) {
		console.error('Error loading page data:', err);
		throw error(500, 'Unable to load page data at the moment. Please try again later.');
	}
};
