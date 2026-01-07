<script lang="ts">
	import { onMount } from 'svelte';
	import '../app.css';
	import { settingsStore } from '$lib/stores/settings';

	// Get commit hash from build-time define
	const commitHash = __COMMIT_HASH__;

	// Load settings on mount
	onMount(() => {
		settingsStore.load();
	});
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<nav class="bg-white dark:bg-gray-800 shadow-sm border-b border-gray-200 dark:border-gray-700">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex justify-between h-16">
				<div class="flex items-center">
					<a href="/" class="text-2xl font-bold text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors">
						nonraid<span class="text-blue-600 dark:text-blue-400">UI</span>
					</a>
				</div>
				<div class="flex items-center space-x-4">
					<a
						href="/settings"
						class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
					>
						<svg class="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
						</svg>
						Settings
					</a>
				</div>
			</div>
		</div>
	</nav>

	<main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
		<slot />
	</main>

	<footer class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 mt-12">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
			<p class="text-center text-sm text-gray-500 dark:text-gray-400">
				Powered by <a href="https://github.com/qvr/nonraid" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline">nonraid</a>.
				Designed by Jean-Ray Arseneau, built by <a href="https://claude.ai" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline">Claude</a>.
				<span class="mx-2">|</span>
				<a href="https://github.com/jrarseneau/nonraid-ui/commit/{commitHash}" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline font-mono">#{commitHash}</a>
				<span class="mx-1">|</span>
				<a href="https://github.com/jrarseneau/nonraid-ui" target="_blank" rel="noopener noreferrer" class="text-blue-600 dark:text-blue-400 hover:underline">GitHub</a>
			</p>
		</div>
	</footer>
</div>
