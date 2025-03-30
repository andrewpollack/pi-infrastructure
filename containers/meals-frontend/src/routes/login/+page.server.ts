import type { Actions, RequestEvent } from './$types';
import { fail, redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const actions: Actions = {
	default: async ({ request, cookies }: RequestEvent) => {
		const formData = await request.formData();

		let res: Response;
		res = await fetch(`${env.API_BASE_URL}/api/login`, {
			method: 'POST',
			body: formData,
			credentials: 'include'
		});

		if (res.status === 401) {
			return fail(401, { message: 'Invalid credentials' });
		}

		const setCookieHeader = res.headers.get('set-cookie');
		if (!setCookieHeader) {
			return fail(401, { message: 'Something went wrong getting cookie' });
		}
		const [cookieNameValue, ...cookieDirectives] = setCookieHeader.split(';');
		const [cookieName, cookieValue] = cookieNameValue.split('=');

		if (!cookieName || !cookieValue) {
			return fail(401, { message: 'Something went wrong setting cookie' });
		}

		cookies.set(cookieName, cookieValue, {
			secure: false,
			httpOnly: true,
			path: '/',
			maxAge: 2.628e6 // 1 month
		});

		throw redirect(303, '/');
	}
};
