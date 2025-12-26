<script lang="ts">
	import { onMount } from 'svelte';
	import { settingsStore } from '$lib/stores/settings';
	import type { Settings } from '$lib/stores/settings';

	type Tab = 'general' | 'notifications';
	let activeTab: Tab = 'general';

	let settings: Settings;
	let loading = true;
	let saving = false;
	let error = '';
	let successMessage = '';

	// Email test state
	let testingEmail = false;
	let testEmailMessage = '';

	// Discord test state
	let testingDiscord = false;
	let testDiscordMessage = '';

	// Handle URL hash for tab navigation
	function updateTabFromHash() {
		const hash = window.location.hash.slice(1);
		if (hash === 'general' || hash === 'notifications') {
			activeTab = hash;
		} else {
			activeTab = 'general';
		}
	}

	function setActiveTab(tab: Tab) {
		activeTab = tab;
		window.location.hash = tab;
	}

	// Load settings on mount
	onMount(async () => {
		await loadSettings();
		updateTabFromHash();

		// Listen for hash changes
		window.addEventListener('hashchange', updateTabFromHash);

		return () => {
			window.removeEventListener('hashchange', updateTabFromHash);
		};
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
		// Validate usage thresholds
		if (settings.thresholds.warning_pct < 0 || settings.thresholds.warning_pct > 100) {
			error = 'Disk usage warning threshold must be between 0 and 100';
			return;
		}
		if (settings.thresholds.critical_pct < 0 || settings.thresholds.critical_pct > 100) {
			error = 'Disk usage critical threshold must be between 0 and 100';
			return;
		}
		if (settings.thresholds.warning_pct >= settings.thresholds.critical_pct) {
			error = 'Disk usage warning threshold must be less than critical threshold';
			return;
		}

		// Validate temperature thresholds
		if (settings.thresholds.temp_warning_c < 0 || settings.thresholds.temp_warning_c > 100) {
			error = 'Temperature warning threshold must be between 0 and 100°C';
			return;
		}
		if (settings.thresholds.temp_critical_c < 0 || settings.thresholds.temp_critical_c > 100) {
			error = 'Temperature critical threshold must be between 0 and 100°C';
			return;
		}
		if (settings.thresholds.temp_warning_c >= settings.thresholds.temp_critical_c) {
			error = 'Temperature warning threshold must be less than critical threshold';
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

	async function testEmail() {
		if (!settings.notifications.email.password) {
			testEmailMessage = 'Please enter your email password in the configuration above';
			return;
		}

		testingEmail = true;
		testEmailMessage = '';

		try {
			const response = await fetch('/api/notifications/test/email', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					config: settings.notifications.email,
					password: settings.notifications.email.password
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			testEmailMessage = '✓ Test email sent successfully!';
		} catch (err) {
			testEmailMessage = `✗ ${err instanceof Error ? err.message : 'Failed to send test email'}`;
		} finally {
			testingEmail = false;
		}
	}

	async function testDiscord() {
		testingDiscord = true;
		testDiscordMessage = '';

		try {
			const response = await fetch('/api/notifications/test/discord', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					webhook_url: settings.notifications.discord.webhook_url
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText);
			}

			testDiscordMessage = '✓ Test Discord notification sent successfully!';
		} catch (err) {
			testDiscordMessage = `✗ ${err instanceof Error ? err.message : 'Failed to send test notification'}`;
		} finally {
			testingDiscord = false;
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

<div class="max-w-5xl mx-auto">
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
			<!-- Tab Navigation -->
			<div class="border-b border-gray-200 dark:border-gray-700">
				<nav class="flex -mb-px" aria-label="Tabs">
					<button
						type="button"
						on:click={() => setActiveTab('general')}
						class="w-1/2 py-4 px-1 text-center border-b-2 font-medium text-sm transition-colors {activeTab === 'general'
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
					>
						General
					</button>
					<button
						type="button"
						on:click={() => setActiveTab('notifications')}
						class="w-1/2 py-4 px-1 text-center border-b-2 font-medium text-sm transition-colors {activeTab === 'notifications'
							? 'border-blue-500 text-blue-600 dark:text-blue-400'
							: 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
					>
						Notifications
					</button>
				</nav>
			</div>

			<form on:submit|preventDefault={saveSettings}>
				<!-- General Tab -->
				{#if activeTab === 'general'}
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
						Thresholds
					</h2>
					<p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
						Configure when disk metrics trigger color warnings and optional notifications.
					</p>

					<!-- Thresholds Table -->
					<div class="overflow-x-auto mb-6">
						<table class="w-full border-collapse">
							<thead>
								<tr class="border-b-2 border-gray-300 dark:border-gray-600">
									<th class="text-left py-3 px-4 text-sm font-semibold text-gray-700 dark:text-gray-300">
										Metric
									</th>
									<th class="text-left py-3 px-4 text-sm font-semibold text-gray-700 dark:text-gray-300">
										Warning
									</th>
									<th class="text-left py-3 px-4 text-sm font-semibold text-gray-700 dark:text-gray-300">
										Critical
									</th>
								</tr>
							</thead>
							<tbody>
								<!-- Disk Usage Row -->
								<tr class="border-b border-gray-200 dark:border-gray-700">
									<td class="py-4 px-4 text-sm text-gray-900 dark:text-white font-medium">
										Disk Usage
									</td>
									<td class="py-4 px-4">
										<div class="flex items-center space-x-2">
											<input
												type="number"
												min="0"
												max="100"
												bind:value={settings.thresholds.warning_pct}
												class="w-20 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500"
											/>
											<span class="text-sm text-gray-600 dark:text-gray-400">%</span>
										</div>
									</td>
									<td class="py-4 px-4">
										<div class="flex items-center space-x-2">
											<input
												type="number"
												min="0"
												max="100"
												bind:value={settings.thresholds.critical_pct}
												class="w-20 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500"
											/>
											<span class="text-sm text-gray-600 dark:text-gray-400">%</span>
										</div>
									</td>
								</tr>
								<!-- Disk Temperature Row -->
								<tr class="border-b border-gray-200 dark:border-gray-700">
									<td class="py-4 px-4 text-sm text-gray-900 dark:text-white font-medium">
										Disk Temperature
									</td>
									<td class="py-4 px-4">
										<div class="flex items-center space-x-2">
											<input
												type="number"
												min="0"
												max="100"
												bind:value={settings.thresholds.temp_warning_c}
												class="w-20 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500"
											/>
											<span class="text-sm text-gray-600 dark:text-gray-400">°C</span>
										</div>
									</td>
									<td class="py-4 px-4">
										<div class="flex items-center space-x-2">
											<input
												type="number"
												min="0"
												max="100"
												bind:value={settings.thresholds.temp_critical_c}
												class="w-20 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500"
											/>
											<span class="text-sm text-gray-600 dark:text-gray-400">°C</span>
										</div>
									</td>
								</tr>
							</tbody>
						</table>
					</div>

					<!-- Color Preview -->
					<div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
						<div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
							Color Preview:
						</div>
						<div class="flex space-x-4">
							<div class="flex-1">
								<div class="text-xs text-gray-600 dark:text-gray-400 mb-1 text-center">
									Normal
								</div>
								<div class="h-7 bg-green-500 dark:bg-green-600 rounded flex items-center justify-center">
									<span class="text-xs font-semibold text-white">Green</span>
								</div>
							</div>
							<div class="flex-1">
								<div class="text-xs text-gray-600 dark:text-gray-400 mb-1 text-center">
									Warning
								</div>
								<div class="h-7 bg-yellow-500 dark:bg-yellow-600 rounded flex items-center justify-center">
									<span class="text-xs font-semibold text-white">Yellow</span>
								</div>
							</div>
							<div class="flex-1">
								<div class="text-xs text-gray-600 dark:text-gray-400 mb-1 text-center">
									Critical
								</div>
								<div class="h-7 bg-red-500 dark:bg-red-600 rounded flex items-center justify-center">
									<span class="text-xs font-semibold text-white">Red</span>
								</div>
							</div>
						</div>
					</div>
					</div>
				{/if}

				<!-- Notifications Tab -->
				{#if activeTab === 'notifications'}
					<!-- Notifications Section -->
					<div class="p-6 border-b border-gray-200 dark:border-gray-700">
					<h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">
						Notifications
					</h2>

					<!-- Enable Notifications Toggle -->
					<label class="flex items-center cursor-pointer mb-6">
						<input
							type="checkbox"
							bind:checked={settings.notifications.enabled}
							class="w-5 h-5 text-blue-600 focus:ring-blue-500 rounded"
						/>
						<span class="ml-3 text-gray-900 dark:text-white font-medium">
							Enable Notifications
						</span>
					</label>

					{#if settings.notifications.enabled}
						<!-- Event Selection -->
						<div class="mb-6">
							<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
								Notify on these events:
							</label>
							<div class="space-y-2">
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.array_health}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Array Health (when array is not healthy)
									</span>
								</label>
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.disk_warning}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Disk Usage Warning (at warning threshold)
									</span>
								</label>
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.disk_critical}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Disk Usage Critical (at critical threshold)
									</span>
								</label>
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.disk_temp_warning}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Disk Temperature Warning (at warning threshold)
									</span>
								</label>
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.disk_temp_critical}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Disk Temperature Critical (at critical threshold)
									</span>
								</label>
								<label class="flex items-center cursor-pointer">
									<input
										type="checkbox"
										bind:checked={settings.notifications.events.disk_status_not_ok}
										class="w-4 h-4 text-blue-600 focus:ring-blue-500 rounded"
									/>
									<span class="ml-3 text-gray-900 dark:text-white">
										Disk Status Not OK (when any disk status is not OK)
									</span>
								</label>
							</div>
						</div>

						<!-- Frequency Selection -->
						<div class="mb-6">
							<label for="frequency" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
								Notification Frequency
							</label>
							<select
								id="frequency"
								bind:value={settings.notifications.frequency}
								class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
							>
								<option value="once">Once (until condition resolved)</option>
								<option value="15m">Every 15 minutes</option>
								<option value="30m">Every 30 minutes</option>
								<option value="1h">Every 1 hour</option>
								<option value="3h">Every 3 hours</option>
								<option value="6h">Every 6 hours</option>
								<option value="12h">Every 12 hours</option>
								<option value="24h">Every 24 hours</option>
							</select>
						</div>

						<!-- Email Service -->
						<div class="mb-6 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<label class="flex items-center cursor-pointer mb-4">
								<input
									type="checkbox"
									bind:checked={settings.notifications.email.enabled}
									class="w-5 h-5 text-blue-600 focus:ring-blue-500 rounded"
								/>
								<span class="ml-3 text-gray-900 dark:text-white font-medium">
									Enable Email Notifications
								</span>
							</label>

							{#if settings.notifications.email.enabled}
								<div class="space-y-4 ml-8">
									<div class="grid grid-cols-2 gap-4">
										<div>
											<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
												SMTP Server
											</label>
											<input
												type="text"
												bind:value={settings.notifications.email.smtp_server}
												placeholder="smtp.gmail.com"
												class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
											/>
										</div>
										<div>
											<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
												Port
											</label>
											<input
												type="number"
												bind:value={settings.notifications.email.smtp_port}
												placeholder="587"
												class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
											/>
										</div>
									</div>

									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
											Username
										</label>
										<input
											type="text"
											bind:value={settings.notifications.email.username}
											placeholder="your-email@example.com"
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
										/>
									</div>

									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
											Password
										</label>
										<input
											type="password"
											bind:value={settings.notifications.email.password}
											placeholder="Enter password"
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
										/>
										<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
											Password stored in plain text. Strongly consider using an App-specific password.
										</p>
									</div>

									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
											From Address
										</label>
										<input
											type="email"
											bind:value={settings.notifications.email.from_address}
											placeholder="noreply@example.com"
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
										/>
									</div>

									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
											To Address
										</label>
										<input
											type="email"
											bind:value={settings.notifications.email.to_address}
											placeholder="your-email@example.com"
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
										/>
									</div>

									<!-- Test Email -->
									<div class="pt-4 border-t border-gray-200 dark:border-gray-600">
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
											Test Email Configuration
										</label>
										<button
											type="button"
											on:click={testEmail}
											disabled={testingEmail}
											class="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-green-400 text-white font-medium rounded-lg transition-colors disabled:cursor-not-allowed"
										>
											{testingEmail ? 'Testing...' : 'Send Test Email'}
										</button>
										{#if testEmailMessage}
											<p class="mt-2 text-sm {testEmailMessage.startsWith('✓') ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
												{testEmailMessage}
											</p>
										{/if}
									</div>
								</div>
							{/if}
						</div>

						<!-- Discord Service -->
						<div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
							<label class="flex items-center cursor-pointer mb-4">
								<input
									type="checkbox"
									bind:checked={settings.notifications.discord.enabled}
									class="w-5 h-5 text-blue-600 focus:ring-blue-500 rounded"
								/>
								<span class="ml-3 text-gray-900 dark:text-white font-medium">
									Enable Discord Notifications
								</span>
							</label>

							{#if settings.notifications.discord.enabled}
								<div class="space-y-4 ml-8">
									<div>
										<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
											Webhook URL
										</label>
										<input
											type="text"
											bind:value={settings.notifications.discord.webhook_url}
											placeholder="https://discord.com/api/webhooks/..."
											class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500"
										/>
									</div>

									<!-- Test Discord -->
									<div class="pt-4 border-t border-gray-200 dark:border-gray-600">
										<button
											type="button"
											on:click={testDiscord}
											disabled={testingDiscord}
											class="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 disabled:bg-indigo-400 text-white font-medium rounded-lg transition-colors disabled:cursor-not-allowed"
										>
											{testingDiscord ? 'Testing...' : 'Test Discord Notification'}
										</button>
										{#if testDiscordMessage}
											<p class="mt-2 text-sm {testDiscordMessage.startsWith('✓') ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
												{testDiscordMessage}
											</p>
										{/if}
									</div>
								</div>
							{/if}
						</div>
					{/if}
					</div>
				{/if}

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
