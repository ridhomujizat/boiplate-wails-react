export namespace app {
	
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
	export class LoginResponse {
	    success: boolean;
	    message: string;
	    token?: string;
	
	    static createFrom(source: any = {}) {
	        return new LoginResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.token = source["token"];
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

}

export namespace dto {
	
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

