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

export async function updateDiskNote(diskId: string, note: string): Promise<void> {
	const response = await fetch(`/api/disks/${diskId}/note`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ note })
	});
	if (!response.ok) {
		const errorText = await response.text();
		throw new Error(errorText || 'Failed to update note');
	}
}

export async function deleteDiskNote(diskId: string): Promise<void> {
	const response = await fetch(`/api/disks/${diskId}/note`, {
		method: 'DELETE'
	});
	if (!response.ok) {
		throw new Error('Failed to delete note');
	}
}
