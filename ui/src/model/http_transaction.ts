import {ExtractQueryString} from "@/model/extract_query";

export interface HttpCookie {
    value?:   string;
    path?:     string;
    domain?:   string;
    expires?:  string;
    // MaxAge=0 means no 'Max-Age' attribute specified.
    // MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
    // MaxAge>0 means Max-Age attribute present and given in seconds
    maxAge?:    number;
    secure?:    boolean;
    httpOnly?:  boolean;
}

export interface SchemaValidationFailure {
    reason?: string;
    location?: string;
    line?: number;
    column?: number;
}

export interface ValidationError {
    message: string;
    reason: string;
    validationType: string;
    validationSubType: string;
    specLine: number;
    specColumn: number;
    howToFix: string;
    schemaValidationErrors?: SchemaValidationFailure[];
    context?: any;
}

export class HttpRequest {
    url?: string;
    method?: string;
    path?: string;
    query?: string;
    headers?: any;
    cookies?: any;
    requestBody?: string;

    constructor() {
        this.headers = {}
        this.cookies = {}
    }

    public extractHeaders(): Map<string, string> {
        return new Map(Object.entries(this.headers));
    }

    public extractQuery(): Map<string, string> {
        return ExtractQueryString(this.query);
    }

    public extractCookies(): Map<string, HttpCookie> {
        return new Map(Object.entries(this.cookies));
    }

    checkContentType(contentType: string): boolean {
        if (this.headers) {
            if (this.headers.has("Content-Type")) {

            }
        }
        return false
    }
}

export class HttpResponse {
    headers?: any;
    cookies?: any;
    statusCode?: number;
    responseBody?: string;

    constructor() {
        this.headers = {}
        this.cookies = {}
    }

    extractHeaders(): Map<string, string> {
        return new Map(Object.entries(this.headers));
    }

    extractCookies(): Map<string, string> {
        return new Map(Object.entries(this.cookies));
    }
}

export interface HttpTransaction {
    timestamp?: number;
    delay?: number;
    httpRequest?: HttpRequest;
    requestValidation?: ValidationError[];
    httpResponse?: HttpResponse;
    responseValidation?: ValidationError[];
    id?: string;
}

export function BuildLiveTransactionFromState(httpTransaction: HttpTransaction): HttpTransaction {
    return {
        delay: httpTransaction.delay,
        timestamp: httpTransaction.timestamp,
        httpRequest: Object.assign(new HttpRequest(), httpTransaction.httpRequest),
        httpResponse: Object.assign(new HttpResponse(), httpTransaction.httpResponse),
        id: httpTransaction.id,
        requestValidation: httpTransaction.requestValidation,
        responseValidation: httpTransaction.responseValidation,
    }
}