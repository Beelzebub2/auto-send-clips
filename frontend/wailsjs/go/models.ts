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
	    monitor_path: string;
	    max_file_size: number;
	    check_interval: number;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.webhook_url = source["webhook_url"];
	        this.monitor_path = source["monitor_path"];
	        this.max_file_size = source["max_file_size"];
	        this.check_interval = source["check_interval"];
	    }
	}

}

