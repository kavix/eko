export namespace api {
	
	export class DiffFile {
	    name: string;
	    original: string;
	    modified: string;
	
	    static createFrom(source: any = {}) {
	        return new DiffFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.original = source["original"];
	        this.modified = source["modified"];
	    }
	}
	export class SnapshotRecord {
	    id: string;
	    message: string;
	    path: string;
	    createdAt: string;
	    filesChanged: number;
	
	    static createFrom(source: any = {}) {
	        return new SnapshotRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.message = source["message"];
	        this.path = source["path"];
	        this.createdAt = source["createdAt"];
	        this.filesChanged = source["filesChanged"];
	    }
	}

}

export namespace main {
	
	export class FrontendSnapshot {
	    id: string;
	    timestamp: string;
	    prompt: string;
	    aiSummary: string;
	    filesChanged: string[];
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new FrontendSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.prompt = source["prompt"];
	        this.aiSummary = source["aiSummary"];
	        this.filesChanged = source["filesChanged"];
	        this.model = source["model"];
	    }
	}

}

