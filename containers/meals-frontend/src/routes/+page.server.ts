import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async () => {
	let emails: string[];
	if (env.EMAILS == null) {
		emails = ['user@example.com'];
	} else {
		emails = env.EMAILS.split(',').map((email) => email.trim());
	}

	const mealsRes = await fetch(`${env.API_BASE_URL}/api/meals`);
	const mealsData = await mealsRes.json();

	const calendarRes = await fetch(`${env.API_BASE_URL}/api/calendar`);
	const allCalendars = await calendarRes.json();

	return {
		allMeals: mealsData.allMeals,
		allCalendars,
		allEmails: emails
	};
};
