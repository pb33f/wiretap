import {HttpTransactionLink} from "@/model/http_transaction";

export interface LinkCacheUpdate {
    transactions: HttpTransactionLink[];
    linkStore: Map<string, Map<string, HttpTransactionLink[]>>;
}

onmessage = function (e: MessageEvent<any>) {
    if (e.data.linkStore) {
        const search: LinkCacheUpdate = e.data;
        const linkStore = search.linkStore;
        const transactions = search.transactions;
        linkStore.forEach((value, key) => {
            const updated = update(key, transactions, value);
            linkStore.set(updated.keyword, updated.links);
        });
        postMessage(linkStore);
    }
}

interface updatedResult {
    keyword: string;
    links: Map<string, HttpTransactionLink[]>;
}
function update(keyword: string,
                transactions: HttpTransactionLink[],
                links: Map<string, HttpTransactionLink[]>): updatedResult {

    // check transactions for keyword
    transactions.forEach((transaction) => {
        const querySegments = transaction.queryString.split('&')
        for (let i = 0; i < querySegments.length; i++) {
            const segment = querySegments[i];
            const keyVal = segment.split('=')
            if (keyVal.length === 2) {
                const key = keyVal[0]
                const val = keyVal[1]
                if (key.toLowerCase() === keyword.toLowerCase()) {
                    if (links) {
                        const existing = links.get(val)
                        if (existing) {
                            let found = false;
                            for (let i = 0; i < existing.length; i++) {
                                if (existing[i].id === transaction.id) {
                                    found = true;
                                }
                            }
                            if (!found) {
                                existing.push(transaction)
                            }
                        } else {
                            links.set(val, [transaction])
                        }
                    } else {
                        links = new Map<string, HttpTransactionLink[]>()
                        links.set(val, [transaction])
                    }
                }
            }
        }
    });

    return {
        keyword: keyword,
        links: links
    }
}