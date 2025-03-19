import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch }) => {
	const response = await fetch(`${env.API_BASE_URL}/api/meals`);

	const data = await response.json();

	return { meals: data.allMeals };
};
