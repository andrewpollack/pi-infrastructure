import type { Cookies } from '@sveltejs/kit';

export function getTokenHeaders(cookies: Cookies): Record<string, string> {
	const token = cookies.get('token');
	if (!token) {
		return {};
	}
	return {
		Cookie: `token=${token ?? ''}`
	};
}
