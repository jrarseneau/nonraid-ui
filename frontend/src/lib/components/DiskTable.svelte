<script lang="ts">
	import type { Disk } from '../types';
	import { getDiskStatusColor, getDiskTypeLabel, formatBytes, formatBytesDetailed } from '../utils';

	export let disks: Disk[];

	// Helper to get display slot (P, Q, or slot number)
	function getDisplaySlot(disk: Disk): string {
		if (disk.type === 'P') return 'P';
		if (disk.type === 'Q') return 'Q';
		return disk.slot.toString();
	}

	// Calculate used space in GiB
	function getUsedSpace(disk: Disk): number {
		if (!disk.filesystem?.usage) return 0;
		const usagePercent = parseInt(disk.filesystem.usage);
		return (disk.size_gb * usagePercent) / 100;
	}

	// Calculate free space in GiB
	function getFreeSpace(disk: Disk): number {
		if (!disk.filesystem?.usage) return disk.size_gb;
		const usagePercent = parseInt(disk.filesystem.usage);
		return (disk.size_gb * (100 - usagePercent)) / 100;
	}

	// Get usage percentage as number
	function getUsagePercent(disk: Disk): number {
		if (!disk.filesystem?.usage) return 0;
		return parseInt(disk.filesystem.usage);
	}

	// Sort disks: parity disks (P, Q) first, then data disks by slot number
	$: sortedDisks = [...disks].sort((a, b) => {
		// P parity always first
		if (a.type === 'P') return -1;
		if (b.type === 'P') return 1;
		// Q parity second
		if (a.type === 'Q') return -1;
		if (b.type === 'Q') return 1;
		// Then sort data disks by slot number
		return a.slot - b.slot;
	});
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
	<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
		<h3 class="text-lg font-bold text-gray-900 dark:text-white">Disks</h3>
	</div>

	<div class="overflow-x-auto">
		<table class="w-full">
			<thead class="bg-gray-50 dark:bg-gray-700/50">
				<tr>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Slot
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Type
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Device
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Name
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Status
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Filesystem
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Disk ID
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Size
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Used
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Free
					</th>
				</tr>
			</thead>
			<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
				{#each sortedDisks as disk}
					<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors">
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm font-medium text-gray-900 dark:text-white">
								{getDisplaySlot(disk)}
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
								{getDiskTypeLabel(disk.type)}
							</span>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900 dark:text-white font-mono">
								{disk.device}
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900 dark:text-white">
								{disk.disk_name || '-'}
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(disk.status)}">
								{disk.status.replace('DISK_', '')}
							</span>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900 dark:text-white">
								{disk.filesystem?.type || '-'}
							</div>
							{#if disk.filesystem?.mountpoint}
								<div class="text-xs text-gray-500 dark:text-gray-400">
									{disk.filesystem.mountpoint}
								</div>
							{/if}
						</td>
						<td class="px-6 py-4">
							<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
								{disk.disk_id}
							</div>
						</td>
						<!-- Size column -->
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900 dark:text-white">
								{formatBytes(disk.size_gb)}
							</div>
						</td>
						<!-- Used column with fuel gauge -->
						<td class="px-6 py-4 whitespace-nowrap">
							{#if disk.filesystem?.usage}
								<div class="relative w-32 h-7 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden">
									<div
										class="absolute inset-0 bg-blue-500 dark:bg-blue-600 transition-all"
										style="width: {getUsagePercent(disk)}%"
									></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-white drop-shadow-md">
											{formatBytesDetailed(getUsedSpace(disk))}
										</span>
									</div>
								</div>
							{:else}
								<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
							{/if}
						</td>
						<!-- Free column with fuel gauge -->
						<td class="px-6 py-4 whitespace-nowrap">
							{#if disk.filesystem?.usage}
								<div class="relative w-32 h-7 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden">
									<div
										class="absolute inset-0 bg-green-500 dark:bg-green-600 transition-all"
										style="width: {100 - getUsagePercent(disk)}%"
									></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-white drop-shadow-md">
											{formatBytesDetailed(getFreeSpace(disk))}
										</span>
									</div>
								</div>
							{:else}
								<div class="relative w-32 h-7 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden">
									<div class="absolute inset-0 bg-green-500 dark:bg-green-600"></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-white drop-shadow-md">
											{formatBytesDetailed(disk.size_gb)}
										</span>
									</div>
								</div>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
