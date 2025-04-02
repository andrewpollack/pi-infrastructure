import type { RequestHandler } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { getTokenHeaders } from '$lib/token-utils';
import type { ExtraItemUpdate } from '$lib/types';

export const POST: RequestHandler = async ({ request, cookies }) => {
	try {
		const updatedOrNewItems: ExtraItemUpdate[] = await request.json();

		const res = await fetch(`${env.API_BASE_URL}/api/items/update`, {
			method: 'POST',
			headers: {
				...getTokenHeaders(cookies),
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(updatedOrNewItems)
		});

		const contentType = res.headers.get('content-type') || '';
		if (!res.ok) {
			if (contentType.includes('application/json')) {
				const errorData = await res.json();
				throw new Error(errorData.error || 'An error occurred while updating meals.');
			} else {
				const errorText = await res.text();
				throw new Error(errorText || 'An error occurred while updating meals.');
			}
		}

		const data = contentType.includes('application/json') ? await res.json() : await res.text();
		return new Response(JSON.stringify(data), {
			status: res.status,
			headers: { 'Content-Type': 'application/json' }
		});
	} catch (error) {
		console.error('Error in update POST handler:', error);
		return new Response(
			JSON.stringify({ error: error instanceof Error ? error.message : 'Request failed' }),
			{ status: 500, headers: { 'Content-Type': 'application/json' } }
		);
	}
};
