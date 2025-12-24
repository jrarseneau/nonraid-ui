import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Settings {
	appearance: 'light' | 'dark' | 'auto';
	thresholds: {
		warning_pct: number;
		critical_pct: number;
	};
}

const defaultSettings: Settings = {
	appearance: 'auto',
	thresholds: {
		warning_pct: 95,
		critical_pct: 98
	}
};

// Create the settings store
function createSettingsStore() {
	const { subscribe, set, update } = writable<Settings>(defaultSettings);

	return {
		subscribe,
		set,
		update,
		// Load settings from API
		load: async () => {
			if (!browser) return;

			try {
				const response = await fetch('/api/settings');
				if (response.ok) {
					const settings = await response.json();
					set(settings);
					applyTheme(settings.appearance);
				}
			} catch (err) {
				console.error('Failed to load settings:', err);
			}
		},
		// Update theme preference
		setAppearance: (appearance: Settings['appearance']) => {
			update(s => ({ ...s, appearance }));
			applyTheme(appearance);
		}
	};
}

// Apply theme to document
function applyTheme(appearance: Settings['appearance']) {
	if (!browser) return;

	const html = document.documentElement;

	if (appearance === 'dark') {
		html.classList.add('dark');
	} else if (appearance === 'light') {
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

// Listen for system theme changes when in auto mode
if (browser) {
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
		// This will be triggered when the system preference changes
		// We'll need to check if we're in auto mode and reapply
		settingsStore.subscribe(settings => {
			if (settings.appearance === 'auto') {
				applyTheme('auto');
			}
		})();
	});
}

export const settingsStore = createSettingsStore();
