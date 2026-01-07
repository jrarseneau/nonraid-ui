export function formatBytes(gib: number): string {
	// Input is in GiB (despite field name "size_gb")
	// Convert GiB to GB (decimal, like drive manufacturers use)
	// 1 GiB = 1.073741824 GB
	const gb = gib * 1.073741824;

	if (gb >= 1000) {
		// Show as TB if >= 1000 GB
		return `${Math.round(gb / 1000)} TB`;
	}
	return `${Math.round(gb)} GB`;
}

export function formatBytesDetailed(gib: number): string {
	// Input is in GiB (despite field name "size_gb")
	// Convert GiB to GB (decimal, like drive manufacturers use)
	// 1 GiB = 1.073741824 GB
	const gb = gib * 1.073741824;

	if (gb >= 1000) {
		// Show as TB with 2 decimal places if >= 1000 GB
		return `${(gb / 1000).toFixed(2)} TB`;
	}
	return `${gb.toFixed(0)} GB`;
}

export function formatDuration(seconds: number): string {
	const hours = Math.floor(seconds / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	const secs = seconds % 60;

	if (hours > 0) {
		return `${hours}h ${minutes}m`;
	}
	if (minutes > 0) {
		return `${minutes}m ${secs}s`;
	}
	return `${secs}s`;
}

export function formatRate(mbps: number): string {
	if (mbps >= 1000) {
		return `${(mbps / 1000).toFixed(2)} GB/s`;
	}
	return `${mbps.toFixed(0)} MB/s`;
}

export function getHealthColor(status: string): string {
	switch (status.toUpperCase()) {
		case 'HEALTHY':
			return 'text-green-600 dark:text-green-400';
		case 'DEGRADED':
			return 'text-yellow-600 dark:text-yellow-400';
		case 'FAILED':
			return 'text-red-600 dark:text-red-400';
		default:
			return 'text-gray-600 dark:text-gray-400';
	}
}

export function getDiskStatusColor(status: string): string {
	switch (status) {
		case 'DISK_OK':
			return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
		case 'DISK_INVALID':
			return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200';
		case 'DISK_MISSING':
			return 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200';
		default:
			return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200';
	}
}

export function getDiskTypeLabel(type: string): string {
	switch (type) {
		case 'P':
			return 'Parity';
		case 'Q':
			return 'Parity 2';
		case 'data':
			return 'Data';
		default:
			return type;
	}
}
