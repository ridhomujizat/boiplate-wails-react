export namespace app {
	
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

}

export namespace dto {
	
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

