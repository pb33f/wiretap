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
        case 103:
            return "Early Hints"
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
        case 418:
            return "I'm a Teapot"
        case 421:
            return "Misdirected Request"
        case 422:
            return "Unprocessable Entity"
        case 423:
            return "Locked"
        case 424:
            return "Failed Dependency"
        case 425:
            return "Too Early"
        case 426:
            return "Upgrade Required"
        case 428:
            return "Precondition Required"
        case 429:
            return "Too Many Requests"
        case 431:
            return "Request Header Fields Too Large"
        case 451:
            return "Unavailable For Legal Reasons"
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

/**
 * Extracts the HTTP code description from the HTTP response.
 *
 * These definitions are taken from https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
 * @param response description of the HTTP response code.
 * @constructor
 */
export function ExtractHTTPCodeDescription(response: HttpResponse): string {
    switch (response.statusCode) {
        case 101:
            return "This code is sent in response to an Upgrade request header from the client and indicates the protocol the server is switching to."
        case 102:
            return "This code indicates that the server has received and is processing the request, but no response is available yet."
        case 103:
            return "This status code is primarily intended to be used with the Link header, letting the user agent start preloading resources while the server prepares a response."
        case 200:
            return "The request succeeded. The result meaning of \"success\" depends on the HTTP method:\n" +
                "\n" +
                "- GET: The resource has been fetched and transmitted in the message body.\n" +
                "- HEAD: The representation headers are included in the response without any message body.\n" +
                "- PUT or POST: The resource describing the result of the action is transmitted in the message body.\n" +
                "- TRACE: The message body contains the request message as received by the server."
        case 201:
            return "The request succeeded, and a new resource was created as a result. This is typically the response sent after POST requests, or some PUT requests."
        case 202:
            return "The request has been received but not yet acted upon. It is noncommittal, since there is no way in HTTP to later send an asynchronous response indicating the outcome of the request. It is intended for cases where another process or server handles the request, or for batch processing."
        case 203:
            return "This response code means the returned metadata is not exactly the same as is available from the origin server, but is collected from a local or a third-party copy. This is mostly used for mirrors or backups of another resource. Except for that specific case, the 200 OK response is preferred to this status."
        case 204:
            return "There is no content to send for this request, but the headers may be useful. The user agent may update its cached headers for this resource with the new ones."
        case 205:
            return "Tells the user agent to reset the document which sent this request."
        case 206:
            return "This response code is used when the Range header is sent from the client to request only part of a resource."
        case 207:
            return "(WebDAV) Conveys information about multiple resources, for situations where multiple status codes might be appropriate."
        case 208:
            return "(WebDAV) Used inside a <dav:propstat> response element to avoid repeatedly enumerating the internal members of multiple bindings to the same collection."
        case 226:
            return "(HTTP Delta Encoding) The server has fulfilled a GET request for the resource, and the response is a representation of the result of one or more instance-manipulations applied to the current instance."
        case 300:
            return "The request has more than one possible response. The user agent or user should choose one of them. (There is no standardized way of choosing one of the responses, but HTML links to the possibilities are recommended so the user can pick.)"
        case 301:
            return "The URL of the requested resource has been changed permanently. The new URL is given in the response."
        case 302:
            return "This response code means that the URI of requested resource has been changed temporarily. Further changes in the URI might be made in the future. Therefore, this same URI should be used by the client in future requests."
        case 303:
            return "The server sent this response to direct the client to get the requested resource at another URI with a GET request."
        case 304:
            return "This is used for caching purposes. It tells the client that the response has not been modified, so the client can continue to use the same cached version of the response."
        case 305:
            return "Defined in a previous version of the HTTP specification to indicate that a requested response must be accessed by a proxy. It has been deprecated due to security concerns regarding in-band configuration of a proxy."
        case 306:
            return "This response code is no longer used; it is just reserved. It was used in a previous version of the HTTP/1.1 specification."
        case 307:
            return "The server sends this response to direct the client to get the requested resource at another URI with the same method that was used in the prior request. This has the same semantics as the 302 Found HTTP response code, with the exception that the user agent must not change the HTTP method used: if a POST was used in the first request, a POST must be used in the second request."
        case 308:
            return "This means that the resource is now permanently located at another URI, specified by the Location: HTTP Response header. This has the same semantics as the 301 Moved Permanently HTTP response code, with the exception that the user agent must not change the HTTP method used: if a POST was used in the first request, a POST must be used in the second request."
        case 400:
            return "The server cannot or will not process the request due to something that is perceived to be a client error (e.g., malformed request syntax, invalid request message framing, or deceptive request routing)."
        case 401:
            return "Although the HTTP standard specifies \"unauthorized\", semantically this response means \"unauthenticated\". That is, the client must authenticate itself to get the requested response."
        case 402:
            return "This response code is reserved for future use. The initial aim for creating this code was using it for digital payment systems, however this status code is used very rarely and no standard convention exists."
        case 403:
            return "The client does not have access rights to the content; that is, it is unauthorized, so the server is refusing to give the requested resource. Unlike 401 Unauthorized, the client's identity is known to the server."
        case 404:
            return "The server cannot find the requested resource. In the browser, this means the URL is not recognized. In an API, this can also mean that the endpoint is valid but the resource itself does not exist. Servers may also send this response instead of 403 Forbidden to hide the existence of a resource from an unauthorized client. This response code is probably the most well known due to its frequent occurrence on the web."
        case 405:
            return "The request method is known by the server but is not supported by the target resource. For example, an API may not allow calling DELETE to remove a resource."
        case 406:
            return "This response is sent when the web server, after performing server-driven content negotiation, doesn't find any content that conforms to the criteria given by the user agent."
        case 407:
            return "This is similar to 401 Unauthorized but authentication is needed to be done by a proxy."
        case 408:
            return "This response is sent on an idle connection by some servers, even without any previous request by the client. It means that the server would like to shut down this unused connection. This response is used much more since some browsers, like Chrome, Firefox 27+, or IE9, use HTTP pre-connection mechanisms to speed up surfing. Also note that some servers merely shut down the connection without sending this message."
        case 409:
            return "This response is sent when a request conflicts with the current state of the server."
        case 410:
            return "This response is sent when the requested content has been permanently deleted from server, with no forwarding address. Clients are expected to remove their caches and links to the resource. The HTTP specification intends this status code to be used for \"limited-time, promotional services\". APIs should not feel compelled to indicate resources that have been deleted with this status code."
        case 411:
            return "Server rejected the request because the Content-Length header field is not defined and the server requires it."
        case 412:
            return "The client has indicated preconditions in its headers which the server does not meet."
        case 413:
            return "Request entity is larger than limits defined by server. The server might close the connection or return an Retry-After header field."
        case 414:
            return "The URI requested by the client is longer than the server is willing to interpret."
        case 415:
            return "The media format of the requested data is not supported by the server, so the server is rejecting the request."
        case 416:
            return "The range specified by the Range header field in the request cannot be fulfilled. It's possible that the range is outside the size of the target URI's data."
        case 417:
            return "This response code means the expectation indicated by the Expect request header field cannot be met by the server."
        case 418:
            return "The server refuses the attempt to brew coffee with a teapot."
        case 421:
            return "The request was directed at a server that is not able to produce a response. This can be sent by a server that is not configured to produce responses for the combination of scheme and authority that are included in the request URI."
        case 422:
            return "(WebDAV) The request was well-formed but was unable to be followed due to semantic errors."
        case 423:
            return "(WebDAV) The resource that is being accessed is locked."
        case 424:
            return "(WebDAV) The request failed due to failure of a previous request."
        case 425:
            return "Indicates that the server is unwilling to risk processing a request that might be replayed."
        case 426:
            return "The server refuses to perform the request using the current protocol but might be willing to do so after the client upgrades to a different protocol."
        case 428:
            return "The origin server requires the request to be conditional. This response is intended to prevent the 'lost update' problem, where a client GETs a resource's state, modifies it, and PUTs it back to the server, when meanwhile a third party has modified the state on the server, leading to a conflict."
        case 429:
            return "The user has sent too many requests in a given amount of time (\"rate limiting\")."
        case 431:
            return "The server is unwilling to process the request because its header fields are too large. The request may be resubmitted after reducing the size of the request header fields."
        case 451:
            return "The user agent requested a resource that cannot legally be provided, such as a web page censored by a government."
        case 500:
            return "The server has encountered a situation it does not know how to handle."
        case 501:
            return "The request method is not supported by the server and cannot be handled. The only methods that servers are required to support (and therefore that must not return this code) are GET and HEAD."
        case 502:
            return "This error response means that the server, while working as a gateway to get a response needed to handle the request, got an invalid response."
        case 503:
            return "The server is not ready to handle the request. Common causes are a server that is down for maintenance or that is overloaded. Note that together with this response, a user-friendly page explaining the problem should be sent. This responses should be used for temporary conditions and the Retry-After: HTTP header should, if possible, contain the estimated time for the recovery of the service."
        case 504:
            return "This error response is given when the server is acting as a gateway and cannot get a response in time."
        case 505:
            return "The HTTP version used in the request is not supported by the server."
        case 506:
            return "The server has an internal configuration error: transparent content negotiation for the request results in a circular reference."
        case 507:
            return "(WebDAV) The method could not be performed on the resource because the server is unable to store the representation needed to successfully complete the request."
        case 508:
            return "(WebDAV) The server detected an infinite loop while processing the request."
        case 510:
            return "Further extensions to the request are required for the server to fulfill it."
        case 511:
            return "Indicates that the client needs to authenticate to gain network access."
        default:
            return "Unknown"
    }
}