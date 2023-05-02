

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

export interface HttpRequest {
    url?: string;
    method?: string;
    path?: string;
    query?: string;
    headers?: Record<string, string>;
    requestBody?: string;
}

export interface HttpResponse {
    httpRequest?: HttpRequest;
    statusCode?: number;
    responseBody?: string;
}

export interface HttpTransaction {
    httpRequest?: HttpRequest;
    requestValidation?: ValidationError[];
    httpResponse?: HttpResponse;
    responseValidation?: ValidationError[];
    id?: string;
}

