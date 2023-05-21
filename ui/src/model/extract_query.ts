export function ExtractQueryString(queryString: string): Map<string, string> {
    const query = new Map<string, string>();
    if (queryString) {
        const queryItems = queryString.split("&");
        for (const item of queryItems) {
            const keyValuePair = item.split("=");
            query.set(decodeURI(keyValuePair[0]), decodeURI(keyValuePair[1]));
        }
    }
    return query;
}