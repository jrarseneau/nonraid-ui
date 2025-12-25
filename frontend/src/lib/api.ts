import type { Status, DiskDetails } from './types';

export async function fetchStatus(): Promise<Status> {
	const response = await fetch('/api/status');
	if (!response.ok) {
		throw new Error(`Failed to fetch status: ${response.statusText}`);
	}
	return response.json();
}

export async function fetchDiskDetails(slot: number): Promise<DiskDetails> {
	const response = await fetch(`/api/disk/${slot}`);
	if (!response.ok) {
		throw new Error(`Failed to fetch disk details: ${response.statusText}`);
	}
	return response.json();
}
