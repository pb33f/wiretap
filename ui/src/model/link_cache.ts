import {Bag, GetBagManager} from "@pb33f/saddlebag";
import {WiretapControls, WiretapFilters} from "@/model/controls";
import {linkCacheFactory} from "@/index";
import {
    WiretapFiltersKey,
    WiretapFiltersStore,
    WiretapHttpTransactionStore, WiretapLinkCacheKey,
    WiretapLinkCacheStore,
    WiretapLocalStorage
} from "@/model/constants";
import {HttpTransaction, HttpTransactionBase, HttpTransactionLink} from "@/model/http_transaction";
import localforage from "localforage";

export class TransactionLinkCache {
    // The key of the outer map is the link keyword that was detected in a transaction.
    // the inner key is *value* of the link keyword that was detected in a transaction (e.g. the value of 'id').
    // the value of the inner map is the list of transactions that contain the link keyword.
    private _state: Map<string, Map<string, HttpTransactionLink[]>>

    private readonly _httpTransactionStore: Bag<HttpTransaction>;
    private readonly _linkCacheStore: Bag<Map<string, Map<string, HttpTransactionLink[]>>>;
    private readonly _filtersStore: Bag<WiretapFilters>;

    private readonly _linkCacheWorker: Worker;
    private _filters: WiretapFilters;

    constructor() {

        this._state = new Map<string, Map<string, HttpTransactionLink[]>>();

        // create a new linkCacheWorker
        this._linkCacheWorker = linkCacheFactory();

        // get transaction store
        this._httpTransactionStore =
            GetBagManager().getBag<HttpTransaction>(WiretapHttpTransactionStore);

        // get link cache store
        this._linkCacheStore =
            GetBagManager().getBag<Map<string, Map<string, HttpTransactionLink[]>>>(WiretapLinkCacheStore);

        // filters store & subscribe to filter changes.
        this._filtersStore = GetBagManager().getBag<WiretapFilters>(WiretapFiltersStore);
        this._filters = this._filtersStore.get(WiretapFiltersKey);
        this._filtersStore.subscribe(WiretapFiltersKey, this.filtersChanged.bind(this))

        // load the link cache from storage
        this.loadLinkCacheFromStorage().then((linkCache) => {
            if (linkCache) {
                this._state = linkCache;
                this._linkCacheStore.set(WiretapLinkCacheKey, this._state);
            } else {
                this._state = new Map<string, Map<string, HttpTransactionLink[]>>();
                this.populateState();
            }
        });
    }

    private populateState() {
        if (this._filters?.filterChain?.length > 0) {
            this._filters.filterChain.forEach((chain) => {
                this._state.set(chain.keyword, new Map<string, HttpTransactionLink[]>())
            });
            this.update().then((result) => {
                this.updated(result)
            }).catch((err) => {
                console.error("it failed", err);
            });
        }
    }

    private async loadLinkCacheFromStorage(): Promise<Map<string, Map<string, HttpTransactionLink[]>>> {
        return localforage.getItem<Map<string, Map<string, HttpTransactionLink[]>>>(WiretapLinkCacheStore);
    }

    private saveLinkCacheToStorage() {
        //console.log('link cache updated', this._state);
        this._linkCacheStore.set(WiretapLinkCacheKey, this._state);
        localforage.setItem(WiretapLinkCacheStore, this._state);
    }

    private filtersChanged(filters: WiretapFilters) {
        this._filters = filters;

        // create new state map with the new filter chain
        const newState = new Map<string, Map<string, HttpTransactionLink[]>>();
        this._filters.filterChain.forEach((chain) => {
            if (!this._state.has(chain.keyword)) {
                newState.set(chain.keyword, new Map<string, HttpTransactionLink[]>())
            } else {
                newState.set(chain.keyword, this._state.get(chain.keyword))
            }
        });

        // set new state.
        this._state = newState;

        // update the link cache
        this.update().then(this.updated.bind(this))
    }

    private updated(state: Map<string, Map<string, HttpTransactionLink[]>>) {
        this._state = state
        this.saveLinkCacheToStorage();
    }

    public sync() {
        this.update().then(this.updated.bind(this))
    }

    public async update(): Promise<Map<string, Map<string, HttpTransactionLink[]>>> {
        const transactions = Array.from(this._httpTransactionStore.export().values())

        // strip out everything from the transactions that we don't need
        const stripped: HttpTransactionLink[] = [];
        transactions.forEach((transaction) => {
            stripped.push({
                id: transaction.id,
                queryString: transaction.httpRequest.query,
            })
        })

        return new Promise((resolve) => {
            this._linkCacheWorker.onmessage = (e) => {
                resolve(e.data)
            }
            this._linkCacheWorker.onerror = (e) => {
                throw new Error(e.message)
            }

            this._linkCacheWorker.postMessage(
                {linkStore: this._state, transactions: stripped})
        });
    }

    findLinks(transaction: HttpTransaction): LinkMatch[] {
        const results: LinkMatch[] = [];
        this._state.forEach((value, parameter) => {
            value.forEach((links, paramValue) => {
                links.forEach((link) => {
                    if (transaction.id === link.id) {
                        results.push({
                            parameter: parameter,
                            value: paramValue,
                            siblings: links
                        });
                    }
                });
            });
        });
        return results;
    }
}

export interface LinkMatch {
    parameter: string;
    value: string;
    siblings: HttpTransactionLink[]
}