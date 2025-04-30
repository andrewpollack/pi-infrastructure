import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ExtraItemsResponse } from '$lib/types';
import { env } from '$env/dynamic/private';
import { getTokenHeaders } from '$lib/token-utils';
import type { AislesResponse } from '$lib/types';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	// Fetch extra items
	const itemsResponse = await fetch(`${env.API_BASE_URL}/api/items`, {
		headers: getTokenHeaders(cookies)
	});

	if (!itemsResponse.ok) {
		if (itemsResponse.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(itemsResponse.status, 'Failed to fetch extra items');
	}

	// Fetch aisles
	const aislesResponse = await fetch(`${env.API_BASE_URL}/api/aisles`, {
		headers: getTokenHeaders(cookies)
	});

	if (!aislesResponse.ok) {
		if (aislesResponse.status === 401) {
			throw redirect(302, '/login');
		}
		throw error(aislesResponse.status, 'Failed to fetch aisles');
	}

	const itemsData: ExtraItemsResponse = await itemsResponse.json();
	const aislesData: AislesResponse = await aislesResponse.json();

	return { 
		extraItems: itemsData.allItems,
		aisles: aislesData.aisles
	};
};
