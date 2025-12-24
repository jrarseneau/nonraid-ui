<script lang="ts">
	import type { Array } from '../types';
	import { getHealthColor, formatBytes } from '../utils';

	export let array: Array;
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h2 class="text-2xl font-bold text-gray-900 dark:text-white">Array: {array.label}</h2>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
				{array.disks_present} of {array.total_slots} disks present
			</p>
		</div>
		<div class="text-right">
			<div class="inline-flex items-center px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-700">
				<span class="text-sm font-medium text-gray-600 dark:text-gray-300 mr-2">State:</span>
				<span class="text-sm font-bold text-gray-900 dark:text-white">{array.state}</span>
			</div>
		</div>
	</div>

	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
		<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Health Status</div>
			<div class="text-2xl font-bold {getHealthColor(array.health.status)}">
				{array.health.status}
			</div>
			{#if array.health.details}
				<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{array.health.details}</div>
			{/if}
		</div>

		<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Total Capacity</div>
			<div class="text-2xl font-bold text-gray-900 dark:text-white">
				{formatBytes(array.size.data_gb)}
			</div>
			<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				{array.size.data_disk_count} data disks
			</div>
		</div>

		<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Parity Protection</div>
			<div class="text-2xl font-bold text-gray-900 dark:text-white">
				{#if array.size.has_second_parity}
					Dual Parity
				{:else if array.size.has_parity}
					Single Parity
				{:else}
					None
				{/if}
			</div>
			{#if array.size.has_parity}
				<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
					{formatBytes(array.size.parity_size_gb)} per parity disk
				</div>
			{/if}
		</div>

		<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4">
			<div class="text-sm font-medium text-gray-500 dark:text-gray-400 mb-1">Disk Issues</div>
			<div class="text-2xl font-bold {array.counters.invalid > 0 || array.counters.missing > 0 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'}">
				{array.counters.invalid + array.counters.missing + array.counters.wrong}
			</div>
			<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
				{array.counters.invalid} invalid, {array.counters.missing} missing
			</div>
		</div>
	</div>
</div>
