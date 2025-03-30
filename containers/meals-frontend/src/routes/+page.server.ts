import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { CalendarResponse, MealsResponse } from '$lib/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const token = cookies.get('token');

	const now = new Date();
	const currentYear = now.getFullYear();
	const currentMonth = now.getMonth() + 1; // JavaScript months are 0-indexed

	const emails = env.EMAILS
		? env.EMAILS.split(',').map((email) => email.trim())
		: ['user@example.com'];

	const mealsRes = await fetch(`${env.API_BASE_URL}/api/meals`, {
		headers: {
			Cookie: `token=${token ?? ''}`
		}
	});
	const calendarRes = await fetch(
		`${env.API_BASE_URL}/api/calendar?year=${currentYear}&month=${currentMonth}`,
		{
			headers: {
				Cookie: `token=${token ?? ''}`
			}
		}
	);
	const extraItemsRes = await fetch(`${env.API_BASE_URL}/api/items`, {
		headers: {
			Cookie: `token=${token ?? ''}`
		}
	});

	if (!mealsRes.ok) {
		if (mealsRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(mealsRes.status, 'Failed to fetch meals');
	}
	if (!calendarRes.ok) {
		if (calendarRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(calendarRes.status, 'Failed to fetch calendar');
	}
	if (!extraItemsRes.ok) {
		if (extraItemsRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(extraItemsRes.status, 'Failed to fetch extra items');
	}

	const mealsData: MealsResponse = await mealsRes.json();
	const calendarData: CalendarResponse = await calendarRes.json();
	const extraItemsData = await extraItemsRes.json();

	return {
		allMeals: mealsData.allMeals,
		currMonthResponse: calendarData.currMonthResponse,
		allEmails: emails,
		allExtraItems: extraItemsData.allItems,
		selectedYear: currentYear,
		selectedMonth: currentMonth
	};
};
