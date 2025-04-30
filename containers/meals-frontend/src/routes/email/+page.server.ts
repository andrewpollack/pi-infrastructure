import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { MealsResponse, EmailsResponse } from '$lib/types';
import { getTokenHeaders } from '$lib/token-utils';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const emailsRes = await fetch(`${env.API_BASE_URL}/api/emails`, {
		headers: getTokenHeaders(cookies)
	});
	const mealsRes = await fetch(`${env.API_BASE_URL}/api/meals`, {
		headers: getTokenHeaders(cookies)
	});
	const extraItemsRes = await fetch(`${env.API_BASE_URL}/api/items`, {
		headers: getTokenHeaders(cookies)
	});

	if (!mealsRes.ok) {
		if (mealsRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(mealsRes.status, 'Failed to fetch meals');
	}
	if (!extraItemsRes.ok) {
		if (extraItemsRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(extraItemsRes.status, 'Failed to fetch extra items');
	}

	if (!emailsRes.ok) {
		if (emailsRes.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(emailsRes.status, 'Failed to fetch emails');
	}

	const emailsData: EmailsResponse = await emailsRes.json();
	const mealsData: MealsResponse = await mealsRes.json();
	const extraItemsData = await extraItemsRes.json();

	return {
		allMeals: mealsData.allMeals,
		allEmails: emailsData.emails,
		allExtraItems: extraItemsData.allItems
	};
};
