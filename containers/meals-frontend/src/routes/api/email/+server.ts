// src/routes/api/email/+server.ts
import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const POST: RequestHandler = async ({ request }) => {
	try {
		const meals = await request.json();

		const res = await fetch(`${env.API_BASE_URL}/api/email`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(meals)
		});

		const data = await res.json();

		return new Response(JSON.stringify(data), {
			status: res.status,
			headers: {
				'Content-Type': 'application/json'
			}
		});
	} catch (error) {
		return new Response(JSON.stringify({ error: 'Request failed' }), {
			status: 500,
			headers: {
				'Content-Type': 'application/json'
			}
		});
	}
};
