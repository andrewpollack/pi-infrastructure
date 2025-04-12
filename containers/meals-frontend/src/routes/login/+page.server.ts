import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { getTokenHeaders } from '$lib/token-utils';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const response = await fetch(`${env.API_BASE_URL}/auth`, {
		headers: getTokenHeaders(cookies)
	});

	if (response.ok) {
		if (response.status === 200) {
			throw redirect(302, '/');
		}
	}

	return { };
};
