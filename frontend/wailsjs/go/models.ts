export namespace main {
	
	export class ThemeMeta {
	    filename: string;
	    name: string;
	    author: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.name = source["name"];
	        this.author = source["author"];
	        this.version = source["version"];
	    }
	}

}

