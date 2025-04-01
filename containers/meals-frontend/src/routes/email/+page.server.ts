import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import type { MealsResponse } from '$lib/types';
import { getTokenHeaders } from '$lib/token-utils';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
    const emails = env.EMAILS
        ? env.EMAILS.split(',').map((email) => email.trim())
        : ['user@example.com'];

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

    const mealsData: MealsResponse = await mealsRes.json();
    const extraItemsData = await extraItemsRes.json();

    return {
        allMeals: mealsData.allMeals,
        allEmails: emails,
        allExtraItems: extraItemsData.allItems
    };
};
