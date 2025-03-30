// +server.ts
import type { RequestHandler } from './$types';
import { env } from '$env/dynamic/private';

export const GET: RequestHandler = async ({ url, cookies }) => {
	const token = cookies.get('token');

	try {
		// Retrieve query parameters and parse them as numbers.
		const yearParam = url.searchParams.get('year');
		const monthParam = url.searchParams.get('month');

		const now = new Date();
		let selectedYear: number = now.getFullYear();
		let selectedMonth: number = now.getMonth() + 1; // JavaScript months are 0-indexed

		if (yearParam) {
			const parsedYear = parseInt(yearParam, 10);
			if (!isNaN(parsedYear)) {
				selectedYear = parsedYear;
			}
		}

		if (monthParam) {
			const parsedMonth = parseInt(monthParam, 10);
			if (!isNaN(parsedMonth) && parsedMonth >= 1 && parsedMonth <= 12) {
				selectedMonth = parsedMonth;
			}
		}

		// Fetch calendar data from the backend API.
		const res = await fetch(
			`${env.API_BASE_URL}/api/calendar?year=${selectedYear}&month=${selectedMonth}`,
			{
				method: 'GET',
				headers: {
					'Content-Type': 'application/json',
					Cookie: `token=${token ?? ''}`
				}
			}
		);

		if (!res.ok) {
			const data = await res.json();
			const message = data.error ? data.error : 'An error occurred while fetching calendar data.';
			return new Response(JSON.stringify({ message }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		// Parse the response data.
		const data = await res.json();
		const currMonthResponse = data.currMonthResponse;
		if (!currMonthResponse) {
			return new Response(JSON.stringify({ message: 'No calendar data found.' }), {
				status: 404,
				headers: { 'Content-Type': 'application/json' }
			});
		}
		// Create a calendar response object.
		const calendarResponse = {
			currMonthResponse,
			selectedYear,
			selectedMonth
		};
		// Return the calendar response.
		return new Response(JSON.stringify(calendarResponse), {
			status: 200,
			headers: {
				'Content-Type': 'application/json'
			}
		});
	} catch (error: any) {
		return new Response(JSON.stringify({ error: error.message || 'Internal Server Error' }), {
			status: 500,
			headers: {
				'Content-Type': 'application/json'
			}
		});
	}
};
