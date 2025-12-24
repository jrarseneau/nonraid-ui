import type { Status } from './types';

export async function fetchStatus(): Promise<Status> {
	const response = await fetch('/api/status');
	if (!response.ok) {
		throw new Error(`Failed to fetch status: ${response.statusText}`);
	}
	return response.json();
}
