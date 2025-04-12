import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { CalendarResponse, MealsResponse, ExtraItemsResponse } from '$lib/types';
import { getTokenHeaders } from '$lib/token-utils';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const now = new Date();
	const currentYear = now.getFullYear();
	const currentMonth = now.getMonth() + 1; // JavaScript months are 0-indexed

	const emails = env.EMAILS
		? env.EMAILS.split(',').map((email) => email.trim())
		: ['user@example.com'];

	const mealsRes = await fetch(`${env.API_BASE_URL}/api/meals`, {
		headers: getTokenHeaders(cookies)
	});
	const calendarRes = await fetch(
		`${env.API_BASE_URL}/api/calendar?year=${currentYear}&month=${currentMonth}`,
		{
			headers: getTokenHeaders(cookies)
		}
	);
	const extraItemsRes = await fetch(`${env.API_BASE_URL}/api/items`, {
		headers: getTokenHeaders(cookies)
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
	const extraItemsData: ExtraItemsResponse = await extraItemsRes.json();
	const extraItems = extraItemsData.allItems.filter((item) => item.Enabled);

	return {
		allMeals: mealsData.allMeals,
		currMonthResponse: calendarData.currMonthResponse,
		allEmails: emails,
		allExtraItems: extraItems,
		selectedYear: currentYear,
		selectedMonth: currentMonth
	};
};
