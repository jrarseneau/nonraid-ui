<script lang="ts">
	import type { Disk } from '../types';
	import { getDiskStatusColor, formatBytes, formatBytesDetailed } from '../utils';
	import { settingsStore } from '$lib/stores/settings';

	export let disks: Disk[];

	// Subscribe to settings for thresholds and view mode
	$: thresholds = $settingsStore.thresholds;
	$: viewMode = $settingsStore.view_mode;
	$: isCondensed = viewMode === 'condensed';

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

	// Get temperature text color based on temperature thresholds
	function getTemperatureColor(temperature: number | null | undefined): string {
		if (temperature === null || temperature === undefined) {
			return 'text-gray-900 dark:text-white';
		}
		if (temperature >= thresholds.temp_critical_c) {
			return 'text-red-600 dark:text-red-400';
		} else if (temperature >= thresholds.temp_warning_c) {
			return 'text-yellow-600 dark:text-yellow-400';
		} else {
			return 'text-gray-900 dark:text-white';
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
	<div class="{isCondensed ? 'px-4 py-2' : 'px-6 py-4'} border-b border-gray-200 dark:border-gray-700">
		<h3 class="{isCondensed ? 'text-base' : 'text-lg'} font-bold text-gray-900 dark:text-white">Disks</h3>
	</div>

	<div class="overflow-x-auto">
		<table class="w-full">
			<thead class="bg-gray-50 dark:bg-gray-700/50">
				<tr>
					<th class="{isCondensed ? 'px-3 py-2 text-xs' : 'px-6 py-3 text-xs'} text-left font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Slot
					</th>
					<th class="{isCondensed ? 'px-3 py-2 text-xs' : 'px-6 py-3 text-xs'} text-left font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Disk ID
					</th>
					<th class="{isCondensed ? 'px-3 py-2 text-xs' : 'px-6 py-3 text-xs'} text-left font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Status
					</th>
					<th class="{isCondensed ? 'px-3 py-2 text-xs' : 'px-6 py-3 text-xs'} text-left font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Temp
					</th>
					{#if !isCondensed}
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
					{:else}
					<th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Size / Usage
					</th>
					{/if}
				</tr>
			</thead>
			<tbody>
				<!-- Parity Section -->
				{#if parityDisks.length > 0}
					<tr class="bg-gray-100 dark:bg-gray-700">
						<td colspan="{isCondensed ? 5 : 8}" class="{isCondensed ? 'px-3 py-1' : 'px-6 py-2'}">
							<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-semibold text-gray-700 dark:text-gray-300">
								Parity
							</div>
						</td>
					</tr>
					{#each parityDisks as disk}
					<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b border-gray-200 dark:border-gray-700">
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-medium text-gray-900 dark:text-white">
								{getDisplaySlot(disk)}
							</div>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'}">
							<div class="flex items-center gap-2">
								<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
									{disk.disk_id}
								</div>
								{#if disk.note}
									<div class="relative group">
										<svg class="{isCondensed ? 'w-3 h-3' : 'w-4 h-4'} text-blue-500 dark:text-blue-400 cursor-help flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
											<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
										</svg>
										<!-- Popover -->
										<div class="absolute hidden group-hover:block z-10 w-80 p-3 bg-gray-800 dark:bg-gray-700 text-white text-sm rounded shadow-lg bottom-full left-1/2 transform -translate-x-1/2 mb-2 whitespace-pre-wrap">
											{disk.note}
											<!-- Arrow -->
											<div class="absolute w-3 h-3 bg-gray-800 dark:bg-gray-700 transform rotate-45 left-1/2 -translate-x-1/2 -bottom-1.5"></div>
										</div>
									</div>
								{/if}
								<a
									href="/disk/{disk.slot}"
									class="flex-shrink-0 text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
									title="View disk details"
								>
									<svg class="{isCondensed ? 'w-3 h-3' : 'w-4 h-4'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</a>
							</div>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(disk.status)}">
								{disk.status.replace('DISK_', '')}
							</span>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							{#if disk.temperature !== undefined && disk.temperature !== null}
								<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-medium {getTemperatureColor(disk.temperature)}">
									{disk.temperature}°C
								</div>
							{:else}
								<div class="{isCondensed ? 'text-xs' : 'text-sm'} text-gray-500 dark:text-gray-400">-</div>
							{/if}
						</td>
						{#if !isCondensed}
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
						{:else}
						<!-- Condensed: combine size/usage into one column -->
						<td class="px-3 py-2 whitespace-nowrap">
							<div class="text-xs text-gray-900 dark:text-white">
								{formatBytes(disk.size_gb)}
							</div>
						</td>
						{/if}
					</tr>
					{/each}
				{/if}

				<!-- Data Section -->
				{#if dataDisks.length > 0}
					<tr class="bg-gray-100 dark:bg-gray-700">
						<td colspan="{isCondensed ? 5 : 8}" class="{isCondensed ? 'px-3 py-1' : 'px-6 py-2'}">
							<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-semibold text-gray-700 dark:text-gray-300">
								Data
							</div>
						</td>
					</tr>
					{#each dataDisks as disk}
					<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b border-gray-200 dark:border-gray-700">
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-medium text-gray-900 dark:text-white">
								{getDisplaySlot(disk)}
							</div>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'}">
							<div class="flex items-center gap-2">
								<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
									{disk.disk_id}
								</div>
								{#if disk.note}
									<div class="relative group">
										<svg class="{isCondensed ? 'w-3 h-3' : 'w-4 h-4'} text-blue-500 dark:text-blue-400 cursor-help flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
											<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
										</svg>
										<!-- Popover -->
										<div class="absolute hidden group-hover:block z-10 w-80 p-3 bg-gray-800 dark:bg-gray-700 text-white text-sm rounded shadow-lg bottom-full left-1/2 transform -translate-x-1/2 mb-2 whitespace-pre-wrap">
											{disk.note}
											<!-- Arrow -->
											<div class="absolute w-3 h-3 bg-gray-800 dark:bg-gray-700 transform rotate-45 left-1/2 -translate-x-1/2 -bottom-1.5"></div>
										</div>
									</div>
								{/if}
								<a
									href="/disk/{disk.slot}"
									class="flex-shrink-0 text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
									title="View disk details"
								>
									<svg class="{isCondensed ? 'w-3 h-3' : 'w-4 h-4'}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								</a>
							</div>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(disk.status)}">
								{disk.status.replace('DISK_', '')}
							</span>
						</td>
						<td class="{isCondensed ? 'px-3 py-2' : 'px-6 py-4'} whitespace-nowrap">
							{#if disk.temperature !== undefined && disk.temperature !== null}
								<div class="{isCondensed ? 'text-xs' : 'text-sm'} font-medium {getTemperatureColor(disk.temperature)}">
									{disk.temperature}°C
								</div>
							{:else}
								<div class="{isCondensed ? 'text-xs' : 'text-sm'} text-gray-500 dark:text-gray-400">-</div>
							{/if}
						</td>
						{#if !isCondensed}
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
						{:else}
						<!-- Condensed: combine size/usage into one column -->
						<td class="px-3 py-2 whitespace-nowrap">
							{#if disk.filesystem?.usage}
								<div class="text-xs text-gray-900 dark:text-white mb-1">
									{formatBytes(disk.size_gb)}
								</div>
								<div class="relative w-24 h-5 bg-gray-200 dark:bg-gray-700 rounded overflow-hidden">
									<div
										class="absolute inset-0 {getGaugeColor(disk)} transition-all"
										style="width: {getUsagePercent(disk)}%"
									></div>
									<div class="absolute inset-0 flex items-center justify-center">
										<span class="text-xs font-semibold text-gray-900 dark:text-white drop-shadow-sm">
											{getUsagePercent(disk)}%
										</span>
									</div>
								</div>
							{:else}
								<div class="text-xs text-gray-900 dark:text-white">
									{formatBytes(disk.size_gb)}
								</div>
							{/if}
						</td>
						{/if}
					</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
</div>
