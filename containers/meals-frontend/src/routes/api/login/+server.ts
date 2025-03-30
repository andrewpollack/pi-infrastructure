import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const POST: RequestHandler = async ({ request, cookies }) => {
	try {
		const login = await request.json();

		const res = await fetch(`${env.API_BASE_URL}/api/login`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(login)
		});

		const data = await res.json();
		if (!res.ok) {
			const message = data.error ? data.error : 'An error occurred while logging in.';
			return new Response(JSON.stringify({ message }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		const setCookieHeader = res.headers.get('set-cookie');
		if (!setCookieHeader) {
			return new Response(JSON.stringify({ message: 'Something went wrong getting cookie' }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		const [cookieNameValue] = setCookieHeader.split(';');
		const [cookieName, cookieValue] = cookieNameValue.split('=');

		if (!cookieName || !cookieValue) {
			return new Response(JSON.stringify({ message: 'Something went wrong setting cookie' }), {
				status: 401,
				headers: { 'Content-Type': 'application/json' }
			});
		}

		cookies.set(cookieName, cookieValue, {
			secure: false,
			httpOnly: true,
			path: '/',
			maxAge: 2.628e6 // 1 month
		});

		return new Response(JSON.stringify(data), {
			status: res.status,
			headers: {
				'Content-Type': 'application/json',
				'Set-Cookie': res.headers.get('set-cookie') || ''
			}
		});
	} catch (error) {
		return new Response(JSON.stringify({ error: `Request failed: ${error}` }), {
			status: 500,
			headers: { 'Content-Type': 'application/json' }
		});
	}
};
