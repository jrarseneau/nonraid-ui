<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fetchStatus } from '$lib/api';
	import type { Status } from '$lib/types';
	import ArrayStatus from '$lib/components/ArrayStatus.svelte';
	import ResyncProgress from '$lib/components/ResyncProgress.svelte';
	import DiskTable from '$lib/components/DiskTable.svelte';

	let status: Status | null = null;
	let error: string | null = null;
	let loading = true;
	let refreshInterval: number;

	async function loadStatus() {
		try {
			status = await fetchStatus();
			error = null;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Unknown error occurred';
			console.error('Failed to fetch status:', e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadStatus();
		// Auto-refresh every 5 seconds
		refreshInterval = setInterval(loadStatus, 5000);
	});

	onDestroy(() => {
		if (refreshInterval) {
			clearInterval(refreshInterval);
		}
	});
</script>

<svelte:head>
	<title>nonraid UI - Dashboard</title>
</svelte:head>

<div class="space-y-6">
	{#if loading && !status}
		<div class="flex items-center justify-center h-64">
			<div class="text-center">
				<div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 dark:border-blue-400"></div>
				<p class="mt-4 text-gray-600 dark:text-gray-400">Loading array status...</p>
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
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Error Loading Status</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">
						<p>{error}</p>
					</div>
					<div class="mt-4">
						<button
							on:click={loadStatus}
							class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-red-700 dark:text-red-200 bg-red-100 dark:bg-red-900/50 hover:bg-red-200 dark:hover:bg-red-900 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
						>
							Retry
						</button>
					</div>
				</div>
			</div>
		</div>
	{:else if status}
		<div class="flex items-center justify-between">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-white">Dashboard</h2>
			<button
				on:click={loadStatus}
				class="inline-flex items-center px-3 py-2 border border-gray-300 dark:border-gray-600 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 dark:text-gray-200 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
			>
				<svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
				</svg>
				Refresh
			</button>
		</div>

		<ArrayStatus array={status.array} disks={status.disks} />

		{#if status.resync?.active}
			<ResyncProgress resync={status.resync} />
		{/if}

		<DiskTable disks={status.disks} />
	{/if}
</div>
