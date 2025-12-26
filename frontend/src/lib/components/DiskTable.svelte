<script lang="ts">
	import type { Disk } from '../types';
	import { getDiskStatusColor, formatBytes, formatBytesDetailed } from '../utils';
	import { settingsStore } from '$lib/stores/settings';

	export let disks: Disk[];

	// Subscribe to settings for thresholds
	$: thresholds = $settingsStore.thresholds;

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

	// Get gauge color based on usage level (uses dynamic thresholds from settings)
	function getGaugeColor(disk: Disk): string {
		const usage = getUsagePercent(disk);
		if (usage >= thresholds.critical_pct) {
			// Critical: red
			return 'bg-red-500 dark:bg-red-600';
		} else if (usage >= thresholds.warning_pct) {
			// Warning: yellow
			return 'bg-yellow-500 dark:bg-yellow-600';
		} else {
			// Normal: green
			return 'bg-green-500 dark:bg-green-600';
		}
	}

	// Separate parity and data disks
	$: parityDisks = disks.filter(d => d.type === 'P' || d.type === 'Q').sort((a, b) => {
		if (a.type === 'P') return -1;
		if (b.type === 'P') return 1;
		return 0;
	});

	$: dataDisks = disks.filter(d => d.type === 'data').sort((a, b) => a.slot - b.slot);
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
						Disk ID
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Status
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Temperature
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Filesystem
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
			<tbody>
				<!-- Parity Section -->
				{#if parityDisks.length > 0}
					<tr class="bg-gray-100 dark:bg-gray-700">
						<td colspan="8" class="px-6 py-2">
							<div class="text-sm font-semibold text-gray-700 dark:text-gray-300">
								Parity
							</div>
						</td>
					</tr>
					{#each parityDisks as disk}
					<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b border-gray-200 dark:border-gray-700">
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm font-medium text-gray-900 dark:text-white">
								{getDisplaySlot(disk)}
							</div>
						</td>
						<td class="px-6 py-4">
							<div class="flex items-center gap-2">
								<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
									{disk.disk_id}
								</div>
								<a
									href="/disk/{disk.slot}"
									class="flex-shrink-0 text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
									title="View disk details"
								>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</a>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(disk.status)}">
								{disk.status.replace('DISK_', '')}
							</span>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							{#if disk.temperature !== undefined && disk.temperature !== null}
								<div class="text-sm text-gray-900 dark:text-white">
									{disk.temperature}°C
								</div>
							{:else}
								<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
							{/if}
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
						<!-- Size column -->
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900 dark:text-white">
								{formatBytes(disk.size_gb)}
							</div>
						</td>
						<!-- Used column - parity disks don't store data -->
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
						</td>
						<!-- Free column - parity disks don't store data -->
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
						</td>
					</tr>
					{/each}
				{/if}

				<!-- Data Section -->
				{#if dataDisks.length > 0}
					<tr class="bg-gray-100 dark:bg-gray-700">
						<td colspan="8" class="px-6 py-2">
							<div class="text-sm font-semibold text-gray-700 dark:text-gray-300">
								Data
							</div>
						</td>
					</tr>
					{#each dataDisks as disk}
					<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b border-gray-200 dark:border-gray-700">
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm font-medium text-gray-900 dark:text-white">
								{getDisplaySlot(disk)}
							</div>
						</td>
						<td class="px-6 py-4">
							<div class="flex items-center gap-2">
								<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
									{disk.disk_id}
								</div>
								<a
									href="/disk/{disk.slot}"
									class="flex-shrink-0 text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
									title="View disk details"
								>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</a>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(disk.status)}">
								{disk.status.replace('DISK_', '')}
							</span>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							{#if disk.temperature !== undefined && disk.temperature !== null}
								<div class="text-sm text-gray-900 dark:text-white">
									{disk.temperature}°C
								</div>
							{:else}
								<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
							{/if}
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
										class="absolute inset-0 {getGaugeColor(disk)} transition-all"
										style="width: {getUsagePercent(disk)}%"
									></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-gray-900 dark:text-white drop-shadow-sm">
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
										class="absolute inset-0 {getGaugeColor(disk)} transition-all"
										style="width: {100 - getUsagePercent(disk)}%"
									></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-gray-900 dark:text-white drop-shadow-sm">
											{formatBytesDetailed(getFreeSpace(disk))}
										</span>
									</div>
								</div>
							{:else}
								<div class="relative w-32 h-7 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden">
									<div class="absolute inset-0 bg-green-500 dark:bg-green-600"></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-gray-900 dark:text-white drop-shadow-sm">
											{formatBytesDetailed(disk.size_gb)}
										</span>
									</div>
								</div>
							{/if}
						</td>
					</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>
