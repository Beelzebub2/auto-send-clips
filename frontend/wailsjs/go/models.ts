export namespace main {
	
	export class AppStatus {
	    uptime: string;
	    isMonitoring: boolean;
	    monitorPath: string;
	    videosSent: number;
	    audiosSent: number;
	    version: string;
	
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
	    }
	}
	export class Config {
	    webhook_url: string;
	    discord_webhook: string;
	    monitor_path: string;
	    max_file_size: number;
	    check_interval: number;
	    audio_extraction: boolean;
	    show_notifications: boolean;
	    auto_compress: boolean;
	
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
	        this.audio_extraction = source["audio_extraction"];
	        this.show_notifications = source["show_notifications"];
	        this.auto_compress = source["auto_compress"];
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

