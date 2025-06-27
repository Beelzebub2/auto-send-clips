export namespace main {
	
	export class AppStatus {
	    uptime: string;
	    isMonitoring: boolean;
	    monitorPath: string;
	    videosSent: number;
	    audiosSent: number;
	    version: string;
	    useMedalTV: boolean;
	    useNVIDIA: boolean;
	    useCustom: boolean;
	    medalTVPath: string;
	    nvidiaPath: string;
	
	    static createFrom(source: any = {}) {
	        return new AppStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uptime = source["uptime"];
	        this.isMonitoring = source["isMonitoring"];
	        this.monitorPath = source["monitorPath"];
	        this.videosSent = source["videosSent"];
	        this.audiosSent = source["audiosSent"];
	        this.version = source["version"];
	        this.useMedalTV = source["useMedalTV"];
	        this.useNVIDIA = source["useNVIDIA"];
	        this.useCustom = source["useCustom"];
	        this.medalTVPath = source["medalTVPath"];
	        this.nvidiaPath = source["nvidiaPath"];
	    }
	}
	export class ClipDisplayData {
	    uuid: string;
	    title: string;
	    gameTitle: string;
	    timeCreated: number;
	    duration: number;
	    thumbnail: string;
	    thumbnailUrl: string;
	    filePath: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ClipDisplayData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uuid = source["uuid"];
	        this.title = source["title"];
	        this.gameTitle = source["gameTitle"];
	        this.timeCreated = source["timeCreated"];
	        this.duration = source["duration"];
	        this.thumbnail = source["thumbnail"];
	        this.thumbnailUrl = source["thumbnailUrl"];
	        this.filePath = source["filePath"];
	        this.status = source["status"];
	    }
	}
	export class Config {
	    webhook_url: string;
	    discord_webhook: string;
	    monitor_path: string;
	    max_file_size: number;
	    check_interval: number;
	    startup_initialization: boolean;
	    windows_startup: boolean;
	    recursive_monitoring: boolean;
	    desktop_shortcut: boolean;
	    use_medaltv_path: boolean;
	    use_nvidia_path: boolean;
	    use_custom_path: boolean;
	    total_clips: number;
	    // Go type: time
	    last_clip_time: any;
	    session_clips: number;
	    total_size_bytes: number;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    last_update_time: any;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.webhook_url = source["webhook_url"];
	        this.discord_webhook = source["discord_webhook"];
	        this.monitor_path = source["monitor_path"];
	        this.max_file_size = source["max_file_size"];
	        this.check_interval = source["check_interval"];
	        this.startup_initialization = source["startup_initialization"];
	        this.windows_startup = source["windows_startup"];
	        this.recursive_monitoring = source["recursive_monitoring"];
	        this.desktop_shortcut = source["desktop_shortcut"];
	        this.use_medaltv_path = source["use_medaltv_path"];
	        this.use_nvidia_path = source["use_nvidia_path"];
	        this.use_custom_path = source["use_custom_path"];
	        this.total_clips = source["total_clips"];
	        this.last_clip_time = this.convertValues(source["last_clip_time"], null);
	        this.session_clips = source["session_clips"];
	        this.total_size_bytes = source["total_size_bytes"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.last_update_time = this.convertValues(source["last_update_time"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Stats {
	    total_clips: number;
	    // Go type: time
	    last_clip_time: any;
	    session_clips: number;
	    total_size_bytes: number;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    last_update_time: any;
	
	    static createFrom(source: any = {}) {
	        return new Stats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_clips = source["total_clips"];
	        this.last_clip_time = this.convertValues(source["last_clip_time"], null);
	        this.session_clips = source["session_clips"];
	        this.total_size_bytes = source["total_size_bytes"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.last_update_time = this.convertValues(source["last_update_time"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace version {
	
	export class BuildInfo {
	    version: string;
	    commit: string;
	    date: string;
	    goVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.date = source["date"];
	        this.goVersion = source["goVersion"];
	    }
	}
	export class UpdateInfo {
	    available: boolean;
	    latestVersion: string;
	    currentVersion: string;
	    releaseURL: string;
	    releaseNotes: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.latestVersion = source["latestVersion"];
	        this.currentVersion = source["currentVersion"];
	        this.releaseURL = source["releaseURL"];
	        this.releaseNotes = source["releaseNotes"];
	        this.error = source["error"];
	    }
	}

}

