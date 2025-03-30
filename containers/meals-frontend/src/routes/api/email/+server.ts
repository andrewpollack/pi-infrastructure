// src/routes/api/email/+server.ts
import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const POST: RequestHandler = async ({ request }) => {
	const cookieHeader = request.headers.get('cookie');
	let token = '';
	if (cookieHeader) {
		const tokenCookie = cookieHeader.split('; ').find((row) => row.startsWith('token='));
		if (tokenCookie) {
			token = tokenCookie.split('=')[1];
		}
	}

	try {
		const meals = await request.json();

		const res = await fetch(`${env.API_BASE_URL}/api/email`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Cookie: `token=${token ?? ''}`
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
		return new Response(JSON.stringify({ error: `Request failed: ${error}` }), {
			status: 500,
			headers: {
				'Content-Type': 'application/json'
			}
		});
	}
};
