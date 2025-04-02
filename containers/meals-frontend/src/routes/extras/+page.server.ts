import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import type { ExtraItemsResponse } from '$lib/types';
import { env } from '$env/dynamic/private';
import { getTokenHeaders } from '$lib/token-utils';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
    const response = await fetch(`${env.API_BASE_URL}/api/items`, {
        headers: getTokenHeaders(cookies)
    });

    if (!response.ok) {
        if (response.status === 401) {
            throw redirect(302, '/login');
        }
        throw error(response.status, 'Failed to fetch meals');
    }

    const data: ExtraItemsResponse = await response.json();

    return { extraItems: data.allItems, };
};
