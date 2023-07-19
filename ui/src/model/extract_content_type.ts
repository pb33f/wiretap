import {HttpRequest, HttpResponse} from "@/model/http_transaction";

export const ContentTypeJSON = "application/json";
export const ContentTypeFormEncoded = "application/x-www-form-urlencoded";
export const ContentTypeXML = "application/xml";

export const ContentTypeMultipartForm = "multipart/form-data";

export const ContentTypeOctetStream = "application/octet-stream";
export const ContentTypeText = "text/plain";
export const ContentTypeHtml = "text/html";


export interface FormDataEntry {
    type: 'file' | 'field';
    name: string;
    filename?: string;
    headers?: Map<string, string[]>;
    value?: string;
}

export interface FormPart {
    type: 'file' | 'field';
    name: string;
    value?: string[];
    headers?: Map<string, string[]>
    files?: FormPart[];
}



export function ExtractContentType(value: string): string {
    return value.split(";")[0];
}

export function IsJsonContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeJSON;
}

export function IsFormEncoded(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeFormEncoded
}


export function IsXmlContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeXML
}

export function IsOctectStreamContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeOctetStream
}

export function IsTextContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeText
}

export function IsHtmlContentType(value: string): boolean {
    const contentType = ExtractContentType(value);
    return contentType === ContentTypeHtml
}

export function ExtractContentTypeFromRequest(request: HttpRequest): string {
    let contentType = request.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = request.headers["Content-Type"];
    if (contentType?.indexOf(";") > 0) {
        contentType = contentType.split(";")[0];
    }
    return contentType;
}

export function ExtractFullContentTypeFromRequest(request: HttpRequest): string {
    let contentType = request.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = request.headers["Content-Type"];
    return contentType;
}


export function ExtractBoundaryFromFormEncodedContentType(contentType: string): string {
    return  contentType.split("boundary=")[1];
}


export function ExtractContentTypeFromResponse(response: HttpResponse): string {
    let contentType = response.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = response.headers["Content-Type"];
    if (contentType && contentType.indexOf(";") > 0) {
        contentType = contentType.split(";")[0];
    }
    return contentType;
}

export function ExtractFullContentTypeFromResponse(response: HttpResponse): string {
    let contentType = response.headers["content-type"];
    if (contentType) {
        return contentType;
    }
    contentType = response.headers["Content-Type"];
    if (contentType && contentType.indexOf(";") > 0) {
        contentType = contentType.split(";")[0];
    }
    return contentType;
}


