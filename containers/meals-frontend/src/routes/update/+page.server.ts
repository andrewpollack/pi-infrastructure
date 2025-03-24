import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { MealsResponse } from '$lib/types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const response = await fetch(`${env.API_BASE_URL}/api/meals`);

		if (!response.ok) {
			throw error(response.status, 'Failed to fetch meals');
		}

		const data: MealsResponse = await response.json();

		return { meals: data.allMeals };
	} catch (err: any) {
		console.error('Error loading page data:', err);
		throw error(500, 'Unable to load page data at the moment. Please try again later.');
	}
};
