<script lang="ts">
	import type { Disk } from '../types';
	import { getDiskStatusColor, getDiskTypeLabel, formatBytes } from '../utils';

	export let disks: Disk[];

	// Helper to get display slot (P, Q, or slot number)
	function getDisplaySlot(disk: Disk): string {
		if (disk.type === 'P') return 'P';
		if (disk.type === 'Q') return 'Q';
		return disk.slot.toString();
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
						Size
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Status
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Filesystem
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Usage
					</th>
					<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						Disk ID
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
							<div class="text-sm text-gray-900 dark:text-white">
								{formatBytes(disk.size_gb)}
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
						<td class="px-6 py-4 whitespace-nowrap">
							{#if disk.filesystem?.usage}
								<div class="flex items-center">
									<div class="text-sm font-medium text-gray-900 dark:text-white mr-2">
										{disk.filesystem.usage}
									</div>
									<div class="w-16 bg-gray-200 dark:bg-gray-600 rounded-full h-2">
										<div
											class="bg-blue-600 dark:bg-blue-500 rounded-full h-2"
											style="width: {disk.filesystem.usage}"
										></div>
									</div>
								</div>
							{:else}
								<div class="text-sm text-gray-500 dark:text-gray-400">-</div>
							{/if}
						</td>
						<td class="px-6 py-4">
							<div class="text-xs text-gray-500 dark:text-gray-400 font-mono max-w-xs truncate" title={disk.disk_id}>
								{disk.disk_id}
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
