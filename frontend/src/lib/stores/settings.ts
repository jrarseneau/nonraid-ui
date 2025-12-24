import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Settings {
	appearance: 'light' | 'dark' | 'auto';
	thresholds: {
		warning_pct: number;
		critical_pct: number;
	};
	notifications: {
		enabled: boolean;
		events: {
			array_health: boolean;
			disk_warning: boolean;
			disk_critical: boolean;
			disk_status_not_ok: boolean;
		};
		frequency: 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
		email: {
			enabled: boolean;
			smtp_server: string;
			smtp_port: number;
			username: string;
			password_md5: string;
			from_address: string;
			to_address: string;
		};
		discord: {
			enabled: boolean;
			webhook_url: string;
		};
	};
}

const defaultSettings: Settings = {
	appearance: 'auto',
	thresholds: {
		warning_pct: 95,
		critical_pct: 98
	},
	notifications: {
		enabled: false,
		events: {
			array_health: false,
			disk_warning: false,
			disk_critical: false,
			disk_status_not_ok: false
		},
		frequency: 'once',
		email: {
			enabled: false,
			smtp_server: '',
			smtp_port: 587,
			username: '',
			password_md5: '',
			from_address: '',
			to_address: ''
		},
		discord: {
			enabled: false,
			webhook_url: ''
		}
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
