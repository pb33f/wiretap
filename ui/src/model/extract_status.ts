import {HttpResponse} from "@/model/http_transaction";

export function ExtractStatusStyleFromCode(response: HttpResponse): string {
    if (response?.statusCode >= 200 && response?.statusCode < 400) {
        return "http200"
    }
    if (response?.statusCode >= 400 && response?.statusCode < 500) {
        return "http400"
    }
    if (response?.statusCode >= 500) {
        return "http500"
    }
    return "pending"
}

export function ExtractHTTPCodeDefinition(response: HttpResponse): string {
    switch (response.statusCode) {
        case 101:
            return "Switching Protocols"
        case 102:
            return "Processing"
        case 200:
            return "OK (Success)"
        case 201:
            return "Created"
        case 202:
            return "Accepted"
        case 203:
            return "Non-Authoritative Information"
        case 204:
            return "No Content"
        case 205:
            return "Reset Content"
        case 206:
            return "Partial Content"
        case 207:
            return "Multi-Status"
        case 208:
            return "Already Reported"
        case 226:
            return "IM Used"
        case 300:
            return "Multiple Choices"
        case 301:
            return "Moved Permanently"
        case 302:
            return "Found"
        case 303:
            return "See Other"
        case 304:
            return "Not Modified"
        case 305:
            return "Use Proxy"
        case 306:
            return "Reserved"
        case 307:
            return "Temporary Redirect"
        case 308:
            return "Permanent Redirect"
        case 400:
            return "Bad Request"
        case 401:
            return "Unauthorized"
        case 402:
            return "Payment Required"
        case 403:
            return "Forbidden"
        case 404:
            return "Not Found"
        case 405:
            return "Method Not Allowed"
        case 406:
            return "Not Acceptable"
        case 407:
            return "Proxy Authentication Required"
        case 408:
            return "Request Timeout"
        case 409:
            return "Conflict"
        case 410:
            return "Gone"
        case 411:
            return "Length Required"
        case 412:
            return "Precondition Failed"
        case 413:
            return "Request Entity Too Large"
        case 414:
            return "Request-URI Too Long"
        case 415:
            return "Unsupported Media Type"
        case 416:
            return "Requested Range Not Satisfiable"
        case 417:
            return "Expectation Failed"
        case 422:
            return "Unprocessable Entity"
        case 423:
            return "Locked"
        case 424:
            return "Failed Dependency"
        case 426:
            return "Upgrade Required"
        case 428:
            return "Precondition Required"
        case 429:
            return "Too Many Requests"
        case 431:
            return "Request Header Fields Too Large"
        case 500:
            return "Internal Server Error"
        case 501:
            return "Not Implemented"
        case 502:
            return "Bad Gateway"
        case 503:
            return "Service Unavailable"
        case 504:
            return "Gateway Timeout"
        case 505:
            return "HTTP Version Not Supported"
        case 506:
            return "Variant Also Negotiates"
        case 507:
            return "Insufficient Storage"
        case 508:
            return "Loop Detected"
        case 510:
            return "Not Extended"
        case 511:
            return "Network Authentication Required"
        default:
            return "Unknown"
    }
}