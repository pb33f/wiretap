import {ExtractQueryString} from "@/model/extract_query";
import {Filter, WiretapFilters} from "@/model/controls";

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

export class HttpTransaction {
    timestamp?: number;
    delay?: number;
    httpRequest?: HttpRequest;
    requestValidation?: ValidationError[];
    httpResponse?: HttpResponse;
    responseValidation?: ValidationError[];
    id?: string;

    constructor(timestamp?: number,
                delay?: number,
                httpRequest?: HttpRequest,
                httpResponse?: HttpResponse,
                id?: string,
                requestValidation?: ValidationError[],
                responseValidation?: ValidationError[]) {
        this.timestamp = timestamp;
        this.delay = delay;
        this.httpRequest = httpRequest;
        this.httpResponse = httpResponse;
        this.id = id;
        this.requestValidation = requestValidation;
        this.responseValidation = responseValidation;
    }

    matchesMethodFilter(filter: WiretapFilters): Filter | boolean {
        if (filter?.filterMethod?.keyword.toLowerCase() === this.httpRequest.method.toLowerCase()) {
            return filter.filterMethod;
        }
        return false;
    }

    matchesKeywordFilter(filter: WiretapFilters): Filter | boolean {
        if (filter?.filterKeywords?.length > 0) {
            for (let i = 0; i < filter.filterKeywords.length; i++) {
                const keywordFilter = filter.filterKeywords[i];
                // check if the keyword filter is in the url.
                if (this.httpRequest.url.toLowerCase().includes(keywordFilter.keyword.toLowerCase())) {
                    return keywordFilter;
                }

                // check if the keyword filter is in the query string.
                if (this.httpRequest.query?.toLowerCase().includes(keywordFilter.keyword.toLowerCase())) {
                    return keywordFilter;
                }

                // check if the keyword filter is in the request body.
                if (this.httpRequest.requestBody?.toLowerCase().includes(keywordFilter.keyword.toLowerCase())) {
                    return keywordFilter;
                }

                // check if the keyword filter is in the response body.
                if (this.httpResponse.responseBody?.toLowerCase().includes(keywordFilter.keyword.toLowerCase())) {
                    return keywordFilter;
                }

                // check headers
                const headers = this.httpRequest.extractHeaders()
                headers.forEach((value) => {
                    if (value.toLowerCase().includes(keywordFilter.keyword.toLowerCase())) {
                        return keywordFilter;
                    }
                });
            }
        }
        return false;
    }

}

export function BuildLiveTransactionFromState(httpTransaction: HttpTransaction): HttpTransaction {
    return new HttpTransaction(
        httpTransaction.timestamp,
        httpTransaction.delay,
        Object.assign(new HttpRequest(), httpTransaction.httpRequest),
        Object.assign(new HttpResponse(), httpTransaction.httpResponse),
        httpTransaction.id,
        httpTransaction.requestValidation,
        httpTransaction.responseValidation);
}