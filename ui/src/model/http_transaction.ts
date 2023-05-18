

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
    specCol: number;
    howToFix: string;
    schemaValidationErrors?: SchemaValidationFailure[];
    context?: any;
}

export class HttpRequest {
    url?: string;
    method?: string;
    path?: string;
    query?: string;
    headers?: Map<string, string>;
    requestBody?: string;

    checkContentType(contentType: string): boolean {
        if (this.headers) {
            if (this.headers.has("Content-Type")) {

            }
        }

        return false
    }
}

export class HttpResponse {
    headers?: Map<string, string>;
    statusCode?: number;
    responseBody?: string;
}

export interface HttpTransaction {
    timestamp?: number;
    httpRequest?: HttpRequest;
    requestValidation?: ValidationError[];
    httpResponse?: HttpResponse;
    responseValidation?: ValidationError[];
    id?: string;
}

