<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { fetchDiskDetails } from '$lib/api';
	import { getDiskStatusColor, formatBytes, formatBytesDetailed } from '$lib/utils';
	import type { DiskDetails } from '$lib/types';

	let diskDetails: DiskDetails | null = null;
	let error: string | null = null;
	let loading = true;

	$: slot = parseInt($page.params.slot);

	async function loadDiskDetails() {
		loading = true;
		try {
			diskDetails = await fetchDiskDetails(slot);
			error = null;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Unknown error occurred';
			console.error('Failed to fetch disk details:', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadDiskDetails();
	});

	// Reload data when slot changes (for next/prev navigation)
	$: if (slot) {
		loadDiskDetails();
	}

	// Helper to get display slot (P, Q, or slot number)
	function getDisplaySlot(type: string, slot: number): string {
		if (type === 'P') return 'P';
		if (type === 'Q') return 'Q';
		return slot.toString();
	}

	// Calculate used space
	function getUsedSpace(): number {
		if (!diskDetails?.disk.filesystem?.usage) return 0;
		const usagePercent = parseInt(diskDetails.disk.filesystem.usage);
		return (diskDetails.disk.size_gb * usagePercent) / 100;
	}

	// Calculate free space
	function getFreeSpace(): number {
		if (!diskDetails?.disk.filesystem?.usage) return diskDetails?.disk.size_gb || 0;
		const usagePercent = parseInt(diskDetails.disk.filesystem.usage);
		return (diskDetails.disk.size_gb * (100 - usagePercent)) / 100;
	}

	// Get usage percentage as number
	function getUsagePercent(): number {
		if (!diskDetails?.disk.filesystem?.usage) return 0;
		return parseInt(diskDetails.disk.filesystem.usage);
	}

	// Find previous and next disks for navigation
	$: prevDisk = diskDetails?.all_disks.find((d, i) => {
		const currentIndex = diskDetails?.all_disks.findIndex(disk => disk.slot === slot);
		return currentIndex !== undefined && i === currentIndex - 1;
	});

	$: nextDisk = diskDetails?.all_disks.find((d, i) => {
		const currentIndex = diskDetails?.all_disks.findIndex(disk => disk.slot === slot);
		return currentIndex !== undefined && i === currentIndex + 1;
	});
</script>

<svelte:head>
	<title>Disk {slot} Details - nonraid UI</title>
</svelte:head>

<div class="space-y-6">
	{#if loading && !diskDetails}
		<div class="flex items-center justify-center h-64">
			<div class="text-center">
				<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400"></div>
				<p class="mt-4 text-gray-600 dark:text-gray-400">Loading disk details...</p>
			</div>
		</div>
	{:else if error}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6">
			<div class="flex items-start">
				<div class="flex-shrink-0">
					<svg class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error Loading Disk Details</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">
						<p>{error}</p>
					</div>
					<div class="mt-4 flex gap-2">
						<a
							href="/"
							class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-red-700 dark:text-red-200 bg-red-100 dark:bg-red-900/50 hover:bg-red-200 dark:hover:bg-red-900"
						>
							Back to Dashboard
						</a>
					</div>
				</div>
			</div>
		</div>
	{:else if diskDetails}
		<!-- Header with Navigation -->
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-4">
				<a
					href="/"
					class="inline-flex items-center text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
				>
					<svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
					Dashboard
				</a>
				<h2 class="text-xl font-semibold text-gray-900 dark:text-white">
					Disk {getDisplaySlot(diskDetails.disk.type, diskDetails.disk.slot)} Details
				</h2>
			</div>

			<!-- Prev/Next Navigation -->
			<div class="flex items-center gap-2">
				{#if prevDisk}
					<a
						href="/disk/{prevDisk.slot}"
						class="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600"
					>
						<svg class="w-4 h-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
						Prev
					</a>
				{/if}
				{#if nextDisk}
					<a
						href="/disk/{nextDisk.slot}"
						class="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600"
					>
						Next
						<svg class="w-4 h-4 ml-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</a>
				{/if}
			</div>
		</div>

		<!-- Disk Overview Card -->
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
				<h3 class="text-lg font-bold text-gray-900 dark:text-white">Disk Overview</h3>
			</div>
			<div class="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
				<!-- Slot -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Slot</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{getDisplaySlot(diskDetails.disk.type, diskDetails.disk.slot)}
					</div>
				</div>

				<!-- Type -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Type</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{diskDetails.disk.type === 'P' || diskDetails.disk.type === 'Q' ? 'Parity' : 'Data'}
					</div>
				</div>

				<!-- Device -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Device</div>
					<div class="mt-1 text-lg font-mono text-gray-900 dark:text-white">
						{diskDetails.disk.device}
					</div>
				</div>

				<!-- Disk ID -->
				<div class="md:col-span-2 lg:col-span-3">
					<div class="text-sm text-gray-500 dark:text-gray-400">Disk ID</div>
					<div class="mt-1 text-sm font-mono text-gray-900 dark:text-white break-all">
						{diskDetails.disk.disk_id}
					</div>
				</div>

				<!-- Status -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Status</div>
					<div class="mt-1">
						<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getDiskStatusColor(diskDetails.disk.status)}">
							{diskDetails.disk.status.replace('DISK_', '')}
						</span>
					</div>
				</div>

				<!-- Temperature -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Temperature</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{#if diskDetails.disk.temperature !== undefined && diskDetails.disk.temperature !== null}
							{diskDetails.disk.temperature}°C
						{:else}
							-
						{/if}
					</div>
				</div>

				<!-- Size -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Total Size</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{formatBytes(diskDetails.disk.size_gb)}
					</div>
				</div>

				{#if diskDetails.disk.filesystem}
					<!-- Filesystem -->
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Filesystem</div>
						<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
							{diskDetails.disk.filesystem.type}
						</div>
					</div>

					<!-- Mount Point -->
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Mount Point</div>
						<div class="mt-1 text-lg font-mono text-gray-900 dark:text-white">
							{diskDetails.disk.filesystem.mountpoint}
						</div>
					</div>

					<!-- Usage -->
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Usage</div>
						<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
							{diskDetails.disk.filesystem.usage}
						</div>
					</div>

					<!-- Used Space -->
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Used Space</div>
						<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
							{formatBytesDetailed(getUsedSpace())} GB
						</div>
					</div>

					<!-- Free Space -->
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Free Space</div>
						<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
							{formatBytesDetailed(getFreeSpace())} GB
						</div>
					</div>
				{/if}

				<!-- Reads -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Total Reads</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{diskDetails.disk.reads.toLocaleString()}
					</div>
				</div>

				<!-- Writes -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Total Writes</div>
					<div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
						{diskDetails.disk.writes.toLocaleString()}
					</div>
				</div>

				<!-- Errors -->
				<div>
					<div class="text-sm text-gray-500 dark:text-gray-400">Errors</div>
					<div class="mt-1 text-lg font-semibold {diskDetails.disk.errors > 0 ? 'text-red-600 dark:text-red-400' : 'text-gray-900 dark:text-white'}">
						{diskDetails.disk.errors}
					</div>
				</div>
			</div>
		</div>

		<!-- SMART Data Section -->
		{#if diskDetails.smart_data}
			<!-- SMART Health Status -->
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-bold text-gray-900 dark:text-white">SMART Health Status</h3>
				</div>
				<div class="p-6">
					<div class="flex items-center gap-4">
						{#if diskDetails.smart_data.smart_status}
							{#if diskDetails.smart_data.smart_status.passed}
								<div class="flex items-center gap-2">
									<svg class="w-8 h-8 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									<span class="text-2xl font-bold text-green-600 dark:text-green-400">PASSED</span>
								</div>
							{:else}
								<div class="flex items-center gap-2">
									<svg class="w-8 h-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									<span class="text-2xl font-bold text-red-600 dark:text-red-400">FAILED</span>
								</div>
							{/if}
						{:else}
							<span class="text-gray-500 dark:text-gray-400">No SMART health status available</span>
						{/if}
					</div>
				</div>
			</div>

			<!-- Device Information -->
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-bold text-gray-900 dark:text-white">Device Information</h3>
				</div>
				<div class="p-6 grid grid-cols-1 md:grid-cols-2 gap-6">
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Device Name</div>
						<div class="mt-1 text-lg font-mono text-gray-900 dark:text-white">
							{diskDetails.smart_data.device.name}
						</div>
					</div>
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Model</div>
						<div class="mt-1 text-lg text-gray-900 dark:text-white">
							{diskDetails.smart_data.device.info_name || '-'}
						</div>
					</div>
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Type</div>
						<div class="mt-1 text-lg text-gray-900 dark:text-white">
							{diskDetails.smart_data.device.type}
						</div>
					</div>
					<div>
						<div class="text-sm text-gray-500 dark:text-gray-400">Protocol</div>
						<div class="mt-1 text-lg text-gray-900 dark:text-white">
							{diskDetails.smart_data.device.protocol}
						</div>
					</div>
				</div>
			</div>

			<!-- Key Metrics -->
			<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
				<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
					<h3 class="text-lg font-bold text-gray-900 dark:text-white">Key Metrics</h3>
				</div>
				<div class="p-6 grid grid-cols-1 md:grid-cols-3 gap-6">
					{#if diskDetails.smart_data.power_on_time}
						<div>
							<div class="text-sm text-gray-500 dark:text-gray-400">Power-On Hours</div>
							<div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
								{diskDetails.smart_data.power_on_time.hours.toLocaleString()}
							</div>
							<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
								≈ {Math.floor(diskDetails.smart_data.power_on_time.hours / 24)} days
							</div>
						</div>
					{/if}
					{#if diskDetails.smart_data.power_cycle_count}
						<div>
							<div class="text-sm text-gray-500 dark:text-gray-400">Power Cycles</div>
							<div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
								{diskDetails.smart_data.power_cycle_count.toLocaleString()}
							</div>
						</div>
					{/if}
					{#if diskDetails.smart_data.temperature}
						<div>
							<div class="text-sm text-gray-500 dark:text-gray-400">Current Temperature</div>
							<div class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">
								{diskDetails.smart_data.temperature.current}°C
							</div>
						</div>
					{/if}
				</div>
			</div>

			<!-- SMART Attributes Table -->
			{#if diskDetails.smart_data.ata_smart_attributes?.table}
				<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg overflow-hidden">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
						<h3 class="text-lg font-bold text-gray-900 dark:text-white">SMART Attributes</h3>
					</div>
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead class="bg-gray-50 dark:bg-gray-700/50">
								<tr>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">ID</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Attribute Name</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Value</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Worst</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Thresh</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Raw Value</th>
									<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Type</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-200 dark:divide-gray-700">
								{#each diskDetails.smart_data.ata_smart_attributes.table as attr}
									<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/30">
										<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">{attr.id}</td>
										<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">{attr.name}</td>
										<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">{attr.value}</td>
										<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">{attr.worst}</td>
										<td class="px-6 py-4 text-sm text-gray-900 dark:text-white">{attr.thresh}</td>
										<td class="px-6 py-4 text-sm font-mono text-gray-900 dark:text-white">{attr.raw.string}</td>
										<td class="px-6 py-4 text-sm">
											{#if attr.flags.prefailure}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400">
													Pre-fail
												</span>
											{:else}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
													Old age
												</span>
											{/if}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{:else}
			<div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-6">
				<p class="text-yellow-800 dark:text-yellow-200">No SMART data available for this disk.</p>
			</div>
		{/if}
	{/if}
</div>
