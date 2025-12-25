export interface Status {
	array: Array;
	resync?: Resync;
	disks: Disk[];
}

export interface Array {
	label: string;
	state: string;
	superblock: string;
	disks_present: number;
	disks_imported: number;
	disks_unassigned: number;
	total_slots: number;
	health: Health;
	size: Size;
	counters: Counters;
	last_sync: LastSync;
}

export interface Health {
	status: string;
	details: string;
	code: number;
}

export interface Size {
	data_gb: number;
	data_disk_count: number;
	has_parity: boolean;
	has_second_parity: boolean;
	parity_size_gb: number;
	second_parity_size_gb: number;
}

export interface Counters {
	missing: number;
	invalid: number;
	wrong: number;
	disabled: number;
	replaced: number;
	new: number;
	sync_errors: number;
	disk_errors: number;
}

export interface LastSync {
	timestamp: number;
	age_seconds: number;
	elapsed_seconds: number;
	status: string;
}

export interface Resync {
	active: boolean;
	paused: boolean;
	pending: boolean;
	action: string;
	progress_percent: number;
	position_gb: number;
	size_gb: number;
	rate_mb_s: number;
	elapsed_seconds: number;
	eta_seconds: number;
}

export interface Disk {
	slot: number;
	type: string;
	size_kb: number;
	size_gb: number;
	device: string;
	status: string;
	errors: number;
	reads: number;
	writes: number;
	disk_id: string;
	disk_name: string;
	filesystem?: Filesystem;
	temperature?: number; // Temperature in Celsius from SMART data
}

export interface Filesystem {
	type: string;
	mountpoint: string;
	usage: string;
}

export interface DiskDetails {
	disk: Disk;
	smart_data?: FullSmartData;
	all_disks: DiskNavigationInfo[];
}

export interface DiskNavigationInfo {
	slot: number;
	type: string;
}

export interface FullSmartData {
	device: DeviceInfo;
	smart_status?: SmartStatus;
	temperature?: {
		current: number;
	};
	power_on_time?: {
		hours: number;
	};
	power_cycle_count?: number;
	ata_smart_attributes?: AtaSmartAttributes;
	smartctl: SmartctlInfo;
}

export interface DeviceInfo {
	name: string;
	info_name: string;
	type: string;
	protocol: string;
}

export interface SmartStatus {
	passed: boolean;
}

export interface AtaSmartAttributes {
	table: SmartAttribute[];
}

export interface SmartAttribute {
	id: number;
	name: string;
	value: number;
	worst: number;
	thresh: number;
	when_failed: string;
	flags: {
		value: number;
		string: string;
		prefailure: boolean;
		updated_online: boolean;
		performance: boolean;
		error_rate: boolean;
		event_count: boolean;
		auto_keep: boolean;
	};
	raw: {
		value: number;
		string: string;
	};
}

export interface SmartctlInfo {
	version: number[];
	exit_status: number;
	messages?: {
		string: string;
		severity: string;
	}[];
}
