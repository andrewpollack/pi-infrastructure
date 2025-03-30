import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { MealsResponse } from '$lib/types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const token = cookies.get('token');

	const response = await fetch(`${env.API_BASE_URL}/api/meals`, {
		headers: {
			Cookie: `token=${token ?? ''}`
		}
	});

	if (!response.ok) {
		if (response.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(response.status, 'Failed to fetch meals');
	}

	const data: MealsResponse = await response.json();

	return { meals: data.allMeals };
};
