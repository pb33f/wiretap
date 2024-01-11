import {customElement, state, query, property} from "lit/decorators.js";
import {html, LitElement} from "lit";
import sharedCss from "@/components/shared.css";
import filtersComponentCss from "./filters.css";
import {ExchangeMethod} from "../../../../../cowboy-components/src/model/exchange_method";
import {WiretapFilters} from "@/model/controls";
import {GlobalDelayChangedEvent} from "@/model/events";
import {SlInput} from "@shoelace-style/shoelace";
import {Bag, BagManager, GetBagManager} from "@pb33f/saddlebag";
import {WiretapFiltersKey, WiretapFiltersStore} from "@/model/constants";
import localforage from "localforage";
import {RanchUtils} from "@pb33f/ranch";

@customElement('wiretap-controls-filters')
export class WiretapControlsFiltersComponent extends LitElement {

    static styles = [sharedCss, filtersComponentCss]

    private _methods: string[] = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD', 'TRACE'];
    private readonly _filtersStore: Bag<WiretapFilters>;
    private readonly _storeManager: BagManager;

    @query('#keyword-input')
    keywordInput: SlInput;

    @query('#chain-input')
    chainInput: SlInput;

    @state()
    private filters: WiretapFilters;

    constructor() {
        super();
        this._storeManager = GetBagManager();
        this._filtersStore = this._storeManager.getBag(WiretapFiltersStore);

        this.loadFiltersFromStorage().then((filters: WiretapFilters) => {
            if (!filters) {
                this.filters = new WiretapFilters();
            } else {
                this.filters = filters;
            }
            this._filtersStore.set(WiretapFiltersKey, this.filters)
        });

    }

    async loadFiltersFromStorage(): Promise<WiretapFilters> {
        return localforage.getItem<WiretapFilters>(WiretapFiltersStore);
    }

    handleKeywordInput() {
        const keyword = this.keywordInput.value;
        this.filters.filterKeywords.push({keyword: keyword, id: RanchUtils.genShortId(5)})
        this.keywordInput.value = ''
        this._filtersStore.set(WiretapFiltersKey, this.filters)
        this.saveFiltersToStorage()
        this.requestUpdate();
    }

    handleChainInput() {
        const param = this.chainInput.value;
        this.filters.filterChain.push({keyword: param, id: RanchUtils.genShortId(5)})
        this.chainInput.value = ''
        this._filtersStore.set(WiretapFiltersKey, this.filters)
        this.saveFiltersToStorage()
        this.requestUpdate();
    }

    private saveFiltersToStorage() {
        localforage.setItem<WiretapFilters>(WiretapFiltersStore, this.filters);
    }

    private removeKeyword(id: string) {
        this.filters.filterKeywords = this.filters.filterKeywords.filter((filter) => {
            return filter.id !== id;
        })
        this._filtersStore.set(WiretapFiltersKey, this.filters)
        this.saveFiltersToStorage()
        this.requestUpdate();
    }

    private removeChain(id: string) {
        this.filters.filterChain = this.filters.filterChain.filter((filter) => {
            return filter.id !== id;
        })
        this._filtersStore.set(WiretapFiltersKey, this.filters)
        this.saveFiltersToStorage()
        this.requestUpdate();
    }

    private methodFilterChanged(value: string) {
        this.filters.filterMethod.keyword = value
        this._filtersStore.set(WiretapFiltersKey, this.filters)
        this.saveFiltersToStorage()
        this.requestUpdate();
    }


    render() {

        // not ready yet, needs a little more UX thinking.
        const requestChainsFeature = html`
            <hr/>
        <h3>Request Chains</h3>
        <p>
            Link related requests together using query parameter keys.
        </p>

        <sl-input @sl-change=${this.handleChainInput} class="label-on-left chain-input" type="text" label="Parameter"
                  help-text="Press enter to add" id="chain-input"></sl-input>

        <div class="chains">
            ${this.filters?.filterChain?.map((filter) => {
                return html`
                        <sl-tag @sl-remove=${(event) => {
                    const tag = event.target;
                    tag.style.opacity = '0';
                    setTimeout(() => {
                        (tag.style.opacity = '1');
                        this.removeChain(filter.id)
                    }, 200);

                }} class="chain" size="small" removable>${filter.keyword}
                        </sl-tag>`
            })}
        </div>`

        return html`
            <sl-select value="${this.filters?.filterMethod.keyword}" class="label-on-left" label="Method" @sl-change=${(change) =>{ this.methodFilterChanged(change.target.value)} } clearable>
                ${this._methods.map((method) => {
                    return html`
                        <sl-option  value="${method}">
                            <sl-tag variant="${ExchangeMethod(method)}" class="method">${method}</sl-tag>
                        </sl-option>`
                })}
            </sl-select>
            
            <hr/>
            <h3>Keyword Filters</h3>
            <p>
                Keywords are matched against the request and response body and
                query parameters.
            </p>

            <sl-input @sl-change=${this.handleKeywordInput} class="label-on-left" type="text" label="Keyword"
                      help-text="Press enter to add" id="keyword-input"  ></sl-input>

            <div class="keywords">
                ${this.filters?.filterKeywords?.map((filter) => {
                    return html`
                        <sl-tag @sl-remove=${(event) => {
                            const tag = event.target;
                            tag.style.opacity = '0';
                            setTimeout(() => {
                                (tag.style.opacity = '1');
                                this.removeKeyword(filter.id)
                            }, 200);

                        }} class="keyword" size="small" removable>${filter.keyword}
                        </sl-tag>`
                })}
            </div>
            
            ${requestChainsFeature}
         
        `
    }
}