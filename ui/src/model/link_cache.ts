import {Bag, GetBagManager} from "@pb33f/saddlebag";
// @ts-ignore
import LinkCacheWorker from "@/workers/link_cache_worker?worker";
import {WiretapFilters} from "@/model/controls";

import {
    WiretapFiltersKey,
    WiretapFiltersStore,
    WiretapHttpTransactionStore, WiretapLinkCacheKey,
    WiretapLinkCacheStore,
} from "@/model/constants";
import {HttpTransaction, HttpTransactionBase, HttpTransactionLink} from "@/model/http_transaction";

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
    private _updateGeneration: number = 0;

    constructor() {

        this._state = new Map<string, Map<string, HttpTransactionLink[]>>();

        // create a new linkCacheWorker
        this._linkCacheWorker = new LinkCacheWorker();

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

        this._linkCacheStore.set(WiretapLinkCacheKey, this._state);
        this.populateState();
    }

    private populateState() {
        if (this._filters?.filterChain?.length > 0) {
            this._filters.filterChain.forEach((chain) => {
                this._state.set(chain.keyword, new Map<string, HttpTransactionLink[]>())
            });
            const generation = this.nextUpdateGeneration();
            this.update().then((result) => {
                this.updated(result, generation)
            }).catch((err) => {
                console.error("it failed", err);
            });
        }
    }

    private saveLinkCache() {
        //console.log('link cache updated', this._state);
        this._linkCacheStore.set(WiretapLinkCacheKey, this._state);
    }

    private filtersChanged(filters: WiretapFilters) {
        this._filters = filters;

        // create new state map with the new filter chain
        const newState = new Map<string, Map<string, HttpTransactionLink[]>>();
        this._filters?.filterChain.forEach((chain) => {
            if (!this._state.has(chain.keyword)) {
                newState.set(chain.keyword, new Map<string, HttpTransactionLink[]>())
            } else {
                newState.set(chain.keyword, this._state.get(chain.keyword))
            }
        });

        // set new state.
        this._state = newState;

        // update the link cache
        const generation = this.nextUpdateGeneration();
        this.update().then((state) => this.updated(state, generation))
    }

    private updated(state: Map<string, Map<string, HttpTransactionLink[]>>, generation?: number) {
        if (generation != undefined && generation !== this._updateGeneration) {
            return;
        }
        this._state = state
        this.saveLinkCache();
    }

    public sync() {
        const generation = this.nextUpdateGeneration();
        this.update().then((state) => this.updated(state, generation))
    }

    public clear() {
        this.nextUpdateGeneration();
        this._state = new Map<string, Map<string, HttpTransactionLink[]>>();
        this._linkCacheStore.set(WiretapLinkCacheKey, this._state);
    }

    private nextUpdateGeneration(): number {
        this._updateGeneration++;
        return this._updateGeneration;
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
        if (transaction) {
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
        }
        return results;
    }
}

export interface LinkMatch {
    parameter: string;
    value: string;
    siblings: HttpTransactionLink[]
}
