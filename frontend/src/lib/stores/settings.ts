import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface Settings {
	appearance: 'light' | 'dark' | 'auto';
	view_mode: 'normal' | 'condensed';
	thresholds: {
		warning_pct: number;
		critical_pct: number;
		temp_warning_c: number;
		temp_critical_c: number;
	};
	notifications: {
		enabled: boolean;
		default_frequency: 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
		events: {
			array_health: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
			disk_warning: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
			disk_critical: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
			disk_temp_warning: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
			disk_temp_critical: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
			disk_status_not_ok: {
				enabled: boolean;
				frequency: 'default' | 'once' | '15m' | '30m' | '1h' | '3h' | '6h' | '12h' | '24h';
			};
		};
		email: {
			enabled: boolean;
			smtp_server: string;
			smtp_port: number;
			username: string;
			password: string;
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
	view_mode: 'normal',
	thresholds: {
		warning_pct: 95,
		critical_pct: 98,
		temp_warning_c: 45,
		temp_critical_c: 55
	},
	notifications: {
		enabled: false,
		default_frequency: 'once',
		events: {
			array_health: { enabled: false, frequency: 'default' },
			disk_warning: { enabled: false, frequency: 'default' },
			disk_critical: { enabled: false, frequency: 'default' },
			disk_temp_warning: { enabled: false, frequency: 'default' },
			disk_temp_critical: { enabled: false, frequency: 'default' },
			disk_status_not_ok: { enabled: false, frequency: 'default' }
		},
		email: {
			enabled: false,
			smtp_server: '',
			smtp_port: 587,
			username: '',
			password: '',
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
