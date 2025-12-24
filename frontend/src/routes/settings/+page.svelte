<script lang="ts">
	import { onMount } from 'svelte';
	import { settingsStore } from '$lib/stores/settings';
	import type { Settings } from '$lib/stores/settings';

	let settings: Settings = {
		appearance: 'auto',
		thresholds: {
			warning_pct: 95,
			critical_pct: 98
		}
	};

	let loading = true;
	let saving = false;
	let error = '';
	let successMessage = '';

	// Load settings on mount
	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		loading = true;
		error = '';
		try {
			const response = await fetch('/api/settings');
			if (!response.ok) {
				throw new Error(`Failed to load settings: ${response.statusText}`);
			}
			settings = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load settings';
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		// Validate thresholds
		if (settings.thresholds.warning_pct < 0 || settings.thresholds.warning_pct > 100) {
			error = 'Warning threshold must be between 0 and 100';
			return;
		}
		if (settings.thresholds.critical_pct < 0 || settings.thresholds.critical_pct > 100) {
			error = 'Critical threshold must be between 0 and 100';
			return;
		}
		if (settings.thresholds.warning_pct >= settings.thresholds.critical_pct) {
			error = 'Warning threshold must be less than critical threshold';
			return;
		}

		saving = true;
		error = '';
		successMessage = '';

		try {
			const response = await fetch('/api/settings', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(settings)
			});

			if (!response.ok) {
				throw new Error(`Failed to save settings: ${response.statusText}`);
			}

			settings = await response.json();

			// Update the global settings store so other components see the changes
			settingsStore.set(settings);

			successMessage = 'Settings saved successfully!';

			// Clear success message after 3 seconds
			setTimeout(() => {
				successMessage = '';
			}, 3000);

			// Apply theme immediately
			applyTheme();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save settings';
		} finally {
			saving = false;
		}
	}

	function applyTheme() {
		const html = document.documentElement;

		if (settings.appearance === 'dark') {
			html.classList.add('dark');
		} else if (settings.appearance === 'light') {
			html.classList.remove('dark');
		} else {
			// Auto mode: use system preference
			if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
				html.classList.add('dark');
			} else {
				html.classList.remove('dark');
			}
		}
	}
</script>

<svelte:head>
	<title>Settings - nonraidUI</title>
</svelte:head>

<div class="max-w-3xl mx-auto">
	<div class="mb-8">
		<h1 class="text-3xl font-bold text-gray-900 dark:text-white">Settings</h1>
		<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
			Configure your nonraidUI preferences
		</p>
	</div>

	{#if loading}
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8 text-center">
			<div class="text-gray-600 dark:text-gray-400">Loading settings...</div>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-800 rounded-lg shadow-lg">
			<form on:submit|preventDefault={saveSettings}>
				<!-- Appearance Section -->
				<div class="p-6 border-b border-gray-200 dark:border-gray-700">
					<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Appearance</h2>
					<div class="space-y-3">
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="appearance"
								value="auto"
								bind:group={settings.appearance}
								class="w-4 h-4 text-blue-600 focus:ring-blue-500"
							/>
							<span class="ml-3 text-gray-900 dark:text-white">
								Auto (follow system preference)
							</span>
						</label>
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="appearance"
								value="light"
								bind:group={settings.appearance}
								class="w-4 h-4 text-blue-600 focus:ring-blue-500"
							/>
							<span class="ml-3 text-gray-900 dark:text-white">Light</span>
						</label>
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="appearance"
								value="dark"
								bind:group={settings.appearance}
								class="w-4 h-4 text-blue-600 focus:ring-blue-500"
							/>
							<span class="ml-3 text-gray-900 dark:text-white">Dark</span>
						</label>
					</div>
				</div>

				<!-- Thresholds Section -->
				<div class="p-6 border-b border-gray-200 dark:border-gray-700">
					<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">
						Disk Usage Thresholds
					</h2>
					<p class="text-sm text-gray-600 dark:text-gray-400 mb-4">
						Configure when disk usage gauges change color based on capacity
					</p>

					<div class="space-y-4">
						<div>
							<label for="warning" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
								Warning Threshold (%)
							</label>
							<div class="flex items-center space-x-3">
								<input
									id="warning"
									type="number"
									min="0"
									max="100"
									bind:value={settings.thresholds.warning_pct}
									class="w-24 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
								/>
								<span class="text-sm text-gray-600 dark:text-gray-400">
									Gauges turn yellow at this usage level
								</span>
							</div>
						</div>

						<div>
							<label for="critical" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
								Critical Threshold (%)
							</label>
							<div class="flex items-center space-x-3">
								<input
									id="critical"
									type="number"
									min="0"
									max="100"
									bind:value={settings.thresholds.critical_pct}
									class="w-24 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
								/>
								<span class="text-sm text-gray-600 dark:text-gray-400">
									Gauges turn red at this usage level
								</span>
							</div>
						</div>

						<!-- Visual preview -->
						<div class="mt-6 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
								Color Preview:
							</div>
							<div class="flex space-x-4">
								<div class="flex-1">
									<div class="text-xs text-gray-600 dark:text-gray-400 mb-1">
										Normal (&lt;{settings.thresholds.warning_pct}%)
									</div>
									<div class="h-7 bg-green-500 dark:bg-green-600 rounded flex items-center justify-center">
										<span class="text-xs font-semibold text-white">Green</span>
									</div>
								</div>
								<div class="flex-1">
									<div class="text-xs text-gray-600 dark:text-gray-400 mb-1">
										Warning ({settings.thresholds.warning_pct}-{settings.thresholds.critical_pct - 1}%)
									</div>
									<div class="h-7 bg-yellow-500 dark:bg-yellow-600 rounded flex items-center justify-center">
										<span class="text-xs font-semibold text-white">Yellow</span>
									</div>
								</div>
								<div class="flex-1">
									<div class="text-xs text-gray-600 dark:text-gray-400 mb-1">
										Critical (≥{settings.thresholds.critical_pct}%)
									</div>
									<div class="h-7 bg-red-500 dark:bg-red-600 rounded flex items-center justify-center">
										<span class="text-xs font-semibold text-white">Red</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Messages -->
				{#if error}
					<div class="p-4 mx-6 mt-6 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 rounded-lg">
						<p class="text-sm text-red-600 dark:text-red-400">{error}</p>
					</div>
				{/if}

				{#if successMessage}
					<div class="p-4 mx-6 mt-6 bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-800 rounded-lg">
						<p class="text-sm text-green-600 dark:text-green-400">{successMessage}</p>
					</div>
				{/if}

				<!-- Action Buttons -->
				<div class="p-6 bg-gray-50 dark:bg-gray-700/50 flex justify-between items-center">
					<a
						href="/"
						class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
					>
						← Back to Dashboard
					</a>
					<button
						type="submit"
						disabled={saving}
						class="px-6 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-medium rounded-lg transition-colors disabled:cursor-not-allowed"
					>
						{saving ? 'Saving...' : 'Save Settings'}
					</button>
				</div>
			</form>
		</div>
	{/if}
</div>
