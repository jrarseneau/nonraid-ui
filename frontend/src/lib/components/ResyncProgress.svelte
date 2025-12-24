<script lang="ts">
	import type { Resync } from '../types';
	import { formatBytes, formatDuration, formatRate } from '../utils';

	export let resync: Resync;
</script>

<div class="bg-gradient-to-r from-blue-500 to-blue-600 dark:from-blue-600 dark:to-blue-700 rounded-lg shadow-lg p-6 text-white">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h3 class="text-lg font-bold">Resync in Progress</h3>
			<p class="text-sm text-blue-100">{resync.action}</p>
		</div>
		<div class="text-right">
			<div class="text-3xl font-bold">{resync.progress_percent}%</div>
			<div class="text-sm text-blue-100">Complete</div>
		</div>
	</div>

	<div class="mb-4">
		<div class="w-full bg-blue-400/30 rounded-full h-3">
			<div
				class="bg-white rounded-full h-3 transition-all duration-300"
				style="width: {resync.progress_percent}%"
			></div>
		</div>
	</div>

	<div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
		<div>
			<div class="text-blue-100">Position</div>
			<div class="font-semibold">{formatBytes(resync.position_gb)} / {formatBytes(resync.size_gb)}</div>
		</div>
		<div>
			<div class="text-blue-100">Speed</div>
			<div class="font-semibold">{formatRate(resync.rate_mb_s)}</div>
		</div>
		<div>
			<div class="text-blue-100">Elapsed</div>
			<div class="font-semibold">{formatDuration(resync.elapsed_seconds)}</div>
		</div>
		<div>
			<div class="text-blue-100">ETA</div>
			<div class="font-semibold">{formatDuration(resync.eta_seconds)}</div>
		</div>
	</div>

	{#if resync.paused}
		<div class="mt-4 bg-yellow-500/20 border border-yellow-300/50 rounded-lg px-4 py-2 text-sm">
			⏸ Resync is currently paused
		</div>
	{/if}
</div>
