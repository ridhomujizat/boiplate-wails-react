export namespace app {
	
	export class ActivitySummary {
	    appName: string;
	    activeTime: number;
	    afkTime: number;
	    sessionCount: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivitySummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.activeTime = source["activeTime"];
	        this.afkTime = source["afkTime"];
	        this.sessionCount = source["sessionCount"];
	        this.percentage = source["percentage"];
	    }
	}
	export class AudioDevice {
	    id: string;
	    name: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioDevice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.type = source["type"];
	    }
	}
	export class DashboardStats {
	    totalActiveTime: number;
	    totalAfkTime: number;
	    totalApps: number;
	    topApp: string;
	
	    static createFrom(source: any = {}) {
	        return new DashboardStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalActiveTime = source["totalActiveTime"];
	        this.totalAfkTime = source["totalAfkTime"];
	        this.totalApps = source["totalApps"];
	        this.topApp = source["topApp"];
	    }
	}
	export class DateRange {
	    startDate: string;
	    endDate: string;
	
	    static createFrom(source: any = {}) {
	        return new DateRange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	    }
	}
	export class PermissionStatus {
	    granted: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new PermissionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.granted = source["granted"];
	        this.message = source["message"];
	    }
	}
	export class RecordingStatus {
	    state: string;
	    duration: number;
	    filePath: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordingStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.state = source["state"];
	        this.duration = source["duration"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	    }
	}
	export class Requirement {
	    id: string;
	    title: string;
	    status: string;
	    progress: number;
	
	    static createFrom(source: any = {}) {
	        return new Requirement(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	    }
	}
	export class SaveSettingsResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SaveSettingsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class StartRecordingResponse {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new StartRecordingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class StopRecordingResponse {
	    success: boolean;
	    message: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new StopRecordingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.filePath = source["filePath"];
	    }
	}
	export class TimelineEvent {
	    id: number;
	    appName: string;
	    windowTitle: string;
	    startTime: string;
	    endTime: string;
	    duration: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TimelineEvent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.appName = source["appName"];
	        this.windowTitle = source["windowTitle"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	        this.duration = source["duration"];
	        this.status = source["status"];
	    }
	}
	export class TopApplication {
	    appName: string;
	    totalDuration: number;
	    sessionCount: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new TopApplication(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appName = source["appName"];
	        this.totalDuration = source["totalDuration"];
	        this.sessionCount = source["sessionCount"];
	        this.percentage = source["percentage"];
	    }
	}

}

export namespace dto {
	
	export class ActivitySettingRequest {
	    pollingInterval: number;
	    afkThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivitySettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pollingInterval = source["pollingInterval"];
	        this.afkThreshold = source["afkThreshold"];
	    }
	}
	export class ActivitySettingResponse {
	    pollingInterval: number;
	    afkThreshold: number;
	
	    static createFrom(source: any = {}) {
	        return new ActivitySettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pollingInterval = source["pollingInterval"];
	        this.afkThreshold = source["afkThreshold"];
	    }
	}
	export class AudioSettingRequest {
	    microphoneId: string;
	    systemAudioEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AudioSettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.microphoneId = source["microphoneId"];
	        this.systemAudioEnabled = source["systemAudioEnabled"];
	    }
	}
	export class AudioSettingResponse {
	    microphoneId: string;
	    systemAudioEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AudioSettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.microphoneId = source["microphoneId"];
	        this.systemAudioEnabled = source["systemAudioEnabled"];
	    }
	}
	export class RecordingSettingRequest {
	    maxRecordingTimeEnabled: boolean;
	    maxRecordingTimeSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new RecordingSettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxRecordingTimeEnabled = source["maxRecordingTimeEnabled"];
	        this.maxRecordingTimeSeconds = source["maxRecordingTimeSeconds"];
	    }
	}
	export class RecordingSettingResponse {
	    maxRecordingTimeEnabled: boolean;
	    maxRecordingTimeSeconds: number;
	
	    static createFrom(source: any = {}) {
	        return new RecordingSettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.maxRecordingTimeEnabled = source["maxRecordingTimeEnabled"];
	        this.maxRecordingTimeSeconds = source["maxRecordingTimeSeconds"];
	    }
	}
	export class SettingRequest {
	    tenantCode: string;
	    baseUrl: string;
	    mqttBroker: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tenantCode = source["tenantCode"];
	        this.baseUrl = source["baseUrl"];
	        this.mqttBroker = source["mqttBroker"];
	    }
	}
	export class SettingResponse {
	    tenantCode: string;
	    baseUrl: string;
	    mqttBroker: string;
	
	    static createFrom(source: any = {}) {
	        return new SettingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tenantCode = source["tenantCode"];
	        this.baseUrl = source["baseUrl"];
	        this.mqttBroker = source["mqttBroker"];
	    }
	}

}

