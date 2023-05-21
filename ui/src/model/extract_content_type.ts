import {HttpRequest, HttpResponse} from "@/model/http_transaction";

export function ExtractContentType(value: string): string {
    return value.split(";")[0];
}

export function IsJsonContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === "application/json";
}

export function IsXmlContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === "application/xml";
}

export function IsOctectStreamContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === "application/octet-stream";
}

export function IsTextContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === "text/plain";
}

export function IsHtmlContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === "text/html";
}

export function ExtractContentTypeFromRequest(request: HttpRequest): string {
    let contentType = request.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = request.headers["Content-Type"];
    return contentType;
}

export function ExtractContentTypeFromResponse(response: HttpResponse): string {
    let contentType = response.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = response.headers["Content-Type"];
    return contentType;
}



