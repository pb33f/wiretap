import {customElement, property, query, state} from "lit/decorators.js";
import {html} from "lit";
import {map} from "lit/directives/map.js";
import {LitElement, TemplateResult} from "lit";
import {BuildLiveTransactionFromState, HttpTransaction} from "@/model/http_transaction";
import transactionViewComponentCss from "./transaction-view.css";
import {KVViewComponent} from "@/components/kv-view/kv-view";
import sharedCss from "@/components/shared.css";
import {SlTab, SlTabGroup} from "@shoelace-style/shoelace";
import {ExtractHTTPCodeDefinition, ExtractStatusStyleFromCode} from "@/model/extract_status";
import {LinkMatch, TransactionLinkCache} from "@/model/link_cache";
import {HttpTransactionItemComponent} from "@/components/transaction/transaction-item";
import {HttpTransactionSelectedEvent, ViolationLocation} from "@/model/events";
import {Bag, GetBagManager} from "@pb33f/saddlebag";
import {WiretapHttpTransactionStore} from "@/model/constants";
import dividerCss from "@/components/divider.css";
import {ResponseBodyViewComponent} from "@/components/transaction/response-body";
import {RequestBodyViewComponent} from "@/components/transaction/request-body";

@customElement('http-transaction-view')
export class HttpTransactionViewComponent extends LitElement {

    static styles = [sharedCss, dividerCss, transactionViewComponentCss];

    @state()
    private _httpTransaction: HttpTransaction

    @query('#violation-tab')
    private _violationTab: SlTab;

    @query('#tabs')
    private _tabs: SlTabGroup;

    private _selectedTab: string;

    @property({type: Boolean})
    hideChain = false;

    private readonly _requestHeadersView: KVViewComponent;
    private readonly _responseHeadersView: KVViewComponent;
    private readonly _requestCookiesView: KVViewComponent;
    private readonly _responseCookiesView: KVViewComponent;
    private readonly _requestQueryView: KVViewComponent;
    private readonly _injectedHeadersView: KVViewComponent;
    private readonly _originalDetailsView: KVViewComponent;

    private _linkCache: TransactionLinkCache;

    // into the matrix.
    private _chainTransactionView: HttpTransactionViewComponent;
    private readonly _httpTransactionStore: Bag<HttpTransaction>;

    @state()
    private _currentLinks: LinkMatch[];

    private _siblings: HttpTransactionItemComponent[];

    constructor() {
        super();
        this._requestHeadersView = new KVViewComponent();
        this._requestCookiesView = new KVViewComponent();
        this._requestCookiesView.keyLabel = 'Cookie Name';
        this._responseHeadersView = new KVViewComponent();
        this._responseCookiesView = new KVViewComponent();
        this._responseCookiesView.keyLabel = 'Cookie Name';
        this._requestQueryView = new KVViewComponent();
        this._injectedHeadersView = new KVViewComponent();
        this._injectedHeadersView.keyLabel = 'Injected Header';
        this._injectedHeadersView.valueLabel = 'Injected Value';
        this._originalDetailsView = new KVViewComponent();
        this._originalDetailsView.keyLabel = 'Detail';
        this._originalDetailsView.valueLabel = 'Mutated Value';


        this._requestQueryView.keyLabel = 'Query Key';
        this._httpTransactionStore =
            GetBagManager().getBag<HttpTransaction>(WiretapHttpTransactionStore);
    }

    set linkCache(value: TransactionLinkCache) {
        this._linkCache = value;
        this.syncLinks();
    }

    get httpTransaction(): HttpTransaction {
        return this._httpTransaction;
    }

    set httpTransaction(value: HttpTransaction) {
        if (value) {
            this._httpTransaction = value;
            if (this._requestHeadersView && value.httpRequest) {
                this._requestHeadersView.data = value.httpRequest.extractHeaders();
                this._requestCookiesView.data = value.httpRequest.extractCookies();
                this._requestQueryView.data = value.httpRequest.extractQuery();
                if (value.httpRequest?.injectedHeaders) {
                    this._injectedHeadersView.data = new Map(Object.entries(value.httpRequest?.injectedHeaders));
                }

                // if there are original details.
                if (value.httpRequest.originalPath != value.httpRequest.path) {
                    this._originalDetailsView.data = new Map([
                        ['Original Path', value.httpRequest.originalPath],
                        ['Rewritten Path', value.httpRequest.path],
                        ['Destination Host',value.httpRequest.host],
                        ['Destination URL', value.httpRequest.url],
                    ]);
                }
            }
            if (this._responseHeadersView && value.httpResponse) {
                this._responseHeadersView.data = value.httpResponse.extractHeaders();
                this._responseCookiesView.data = value.httpResponse.extractCookies();
            }
        } else {
            this._httpTransaction = null;
            this._requestCookiesView.data = null;
            this._requestHeadersView.data = null;
            this._requestQueryView.data = null;
            this._responseCookiesView.data = null;
            this._responseHeadersView.data = null;
        }
        this.syncLinks()
        if (this._chainTransactionView) {
            this._chainTransactionView = null
        }

    }

    tabSelected(event: CustomEvent) {
        this._selectedTab = event.detail.name;
    }


    render() {

        if (this._httpTransaction) {

            const req = this._httpTransaction?.httpRequest;
            const resp = this._httpTransaction?.httpResponse;

            const requestViolations: TemplateResult = html`
                ${this._httpTransaction?.requestValidation?.length > 0 ? html`<h3>Request Violations</h3>` : html``}
                ${map(this._httpTransaction.requestValidation, (i) => {
                    return html`
                        <wiretap-violation-view .violation="${i}"></wiretap-violation-view>
                    `
                })}`;

            const responseViolations: TemplateResult = html`
                ${this._httpTransaction?.responseValidation?.length > 0 ? html`<h3>Response Violations</h3>` : html``}
                ${map(this._httpTransaction.responseValidation, (i) => {
                    return html`
                        <wiretap-violation-view .violation="${i}"></wiretap-violation-view>
                    `
                })}`;


            let total = 0;
            let violations: TemplateResult = html`Violations`;
            if (this._httpTransaction?.requestValidation?.length > 0 || this._httpTransaction?.responseValidation?.length > 0) {

                if (this._httpTransaction?.requestValidation?.length > 0) {
                    total += this._httpTransaction.requestValidation.length;
                }
                if (this._httpTransaction?.responseValidation?.length > 0) {
                    total += this._httpTransaction.responseValidation.length;
                }
                violations = html`Violations
                <sl-badge variant="warning" class="violation-badge">${total}</sl-badge>`;
            }
            const noData: TemplateResult = html`
                <div class="empty-data ok">
                    <sl-icon name="patch-check" class="ok-icon"></sl-icon>
                    <br/>
                    API call is compliant
                </div>`;


            const responseBodyView = new ResponseBodyViewComponent(resp);
            const requestBodyView = new RequestBodyViewComponent(req);

            let requestTab: TemplateResult;
            if (req.requestBody == null || req.requestBody.length <= 0) {
                requestTab = html`
                    <sl-tab slot="nav" panel="request-body" class="tab-secondary" disabled>Body</sl-tab>`;
            } else {
                requestTab = html`
                    <sl-tab slot="nav" panel="request-body" class="tab-secondary">Body</sl-tab>`;
            }


            let originalTab: TemplateResult;
            if (req.injectedHeaders?.size > 0 || req.droppedHeaders?.length > 0 || req.path != req.originalPath) {
                originalTab = html`
                    <sl-tab slot="nav" panel="request-origins" class="tab-secondary">Mutations</sl-tab>`;
            } else {
                originalTab = html`
                    <sl-tab slot="nav" panel="request-origins" class="tab-secondary" disabled>Mutations</sl-tab>`;
            }

            let droppedHeaders: TemplateResult;
            if (req.droppedHeaders.length > 0) {
                droppedHeaders = html`
                    <h3>Dropped Headers</h3>
                    <ul>
                        ${req.droppedHeaders.map((h) => {
                            return html`
                                <li class="dropped-header">${h}</li>`
                        })}
                    </ul>
                `
            }

            const tabGroup: TemplateResult = html`
                <sl-tab-group id="tabs" @sl-tab-show=${this.tabSelected}>
                    <sl-tab slot="nav" panel="violations" id="violation-tab" class="tab">${violations}</sl-tab>
                    <sl-tab slot="nav" panel="request" class="tab">Request</sl-tab>
                    <sl-tab slot="nav" panel="response" class="tab">Response</sl-tab>
                    ${this._currentLinks?.length > 0 ? html`
                        <sl-tab slot="nav" panel="chain" class="tab">Chain</sl-tab>` : null}
                    <sl-tab-panel name="violations" class="tab-panel">
                        ${total <= 0 ? noData : null}
                        ${requestViolations}
                        ${(this._httpTransaction?.requestValidation?.length > 0
                                && this._httpTransaction?.responseValidation?.length > 0) ? html`
                            <hr/>` : null}
                        ${responseViolations}
                    </sl-tab-panel>
                    <sl-tab-panel name="request">
                        <sl-tab-group class="secondary-tabs" placement="start">
                            ${requestTab}
                            <sl-tab slot="nav" panel="request-query" class="tab">Query</sl-tab>
                            <sl-tab slot="nav" panel="request-headers" class="tab-secondary">Headers</sl-tab>
                            <sl-tab slot="nav" panel="request-cookies" class="tab-secondary">Cookies</sl-tab>
                            ${originalTab}
                            <sl-tab-panel name="request-headers">
                                ${this._requestHeadersView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-cookies">
                                ${this._requestCookiesView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-body">
                                ${requestBodyView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-query">
                                ${this._requestQueryView}
                            </sl-tab-panel>

                            <sl-tab-panel name="request-origins">
                                ${req.injectedHeaders != null ? html`<h3>Injected
                                    headers</h3>${this._injectedHeadersView}
                                <hr/>` : null}
                                ${req.originalPath != req.path ? html`<h3>Destination details</h3>${this._originalDetailsView}
                                <hr/>` : null}
                                ${droppedHeaders}
                            </sl-tab-panel>
                        </sl-tab-group>
                    </sl-tab-panel>
                    <sl-tab-panel name="response">
                        <sl-tab-group class="secondary-tabs" placement="start">
                            <sl-tab slot="nav" panel="response-body" class="tab-secondary">Body</sl-tab>
                            <sl-tab slot="nav" panel="response-code" class="tab-secondary">Code</sl-tab>
                            <sl-tab slot="nav" panel="response-headers" class="tab-secondary">Headers</sl-tab>
                            <sl-tab slot="nav" panel="response-cookies" class="tab-secondary">Cookies</sl-tab>
                            <sl-tab-panel name="response-code">
                                <h2 class="${ExtractStatusStyleFromCode(resp)}">${resp.statusCode}</h2>
                                <p class="response-code">${ExtractHTTPCodeDefinition(resp)}</p>
                            </sl-tab-panel>
                            <sl-tab-panel name="response-headers">
                                ${this._responseHeadersView}
                            </sl-tab-panel>
                            <sl-tab-panel name="response-cookies">
                                ${this._responseCookiesView}
                            </sl-tab-panel>
                            <sl-tab-panel name="response-body">
                                ${responseBodyView}
                            </sl-tab-panel>
                        </sl-tab-group>
                    </sl-tab-panel>
                    ${this._currentLinks?.length > 0 ? this.renderChainTabPanel() : null}
                </sl-tab-group>`

            return html`${tabGroup}`
        } else {

            return html`
                <div class="empty-data engage">
                    <sl-icon name="arrow-up-square" class="up-icon"></sl-icon>
                    <br/>
                    Select an API call to explore...
                </div>`
        }
    }

    chainTransactionSelected(event: CustomEvent) {
        if (!this._chainTransactionView) {
            this._chainTransactionView = new HttpTransactionViewComponent()
        }
        this._chainTransactionView.httpTransaction = event.detail;
        this._chainTransactionView.linkCache = this._linkCache;
        this._chainTransactionView.hideChain = true;
        this._chainTransactionView._currentLinks = [];
        this.requestUpdate();
    }

    renderLinkMatch(linkMatch: LinkMatch, hideKv: boolean = false): TemplateResult {

        const paramKVComponent: KVViewComponent = new KVViewComponent();
        const paramData = new Map<string, string>();
        paramData.set(linkMatch.parameter, linkMatch.value);
        paramKVComponent.data = paramData;
        paramKVComponent.keyLabel = "Query Parameter";
        paramKVComponent.valueLabel = "Value";

        const siblings: HttpTransactionItemComponent[] = [];
        const timelineItems: TemplateResult[] = [];

        let tsDiff = 0;
        let tsTotal = 0;
        linkMatch.siblings.forEach((sibling) => {
            const transaction = BuildLiveTransactionFromState(this._httpTransactionStore.get(sibling.id));
            //if (transaction.id !== this._httpTransaction.id) {
            const siblingComponent = new HttpTransactionItemComponent(transaction, this._linkCache);
            siblingComponent.hideControls = true;
            if (this._chainTransactionView && this._chainTransactionView.httpTransaction?.id == transaction.id) {
                siblingComponent.setActive();
            } else {
                siblingComponent.disable();
            }
            siblingComponent.addEventListener(HttpTransactionSelectedEvent, this.chainTransactionSelected.bind(this));
            siblings.push(siblingComponent);

            let tsFormat = "";
            let thisDiff = transaction.httpRequest.timestamp - tsDiff;
            if (thisDiff != transaction.httpRequest.timestamp) {
                if (thisDiff > 1000) {
                    tsFormat = "+" + (thisDiff / 1000).toFixed(3) + "s";
                }
                if (thisDiff > 60000) {
                    tsFormat = "+" + ((thisDiff / 1000) / 60).toFixed(2) + "m";
                }
                if (thisDiff <= 1000) {
                    tsFormat = "+" + thisDiff + "ms";
                }

                tsTotal = tsTotal + thisDiff + (transaction.httpResponse?.timestamp - transaction.httpRequest?.timestamp);
            }

            let linkIcon = 'link-45deg';
            let linkCss = 'color: var(--dark-font-color);';
            let tsCss = '';
            if (this._httpTransaction.id == transaction.id) {
                linkIcon = 'arrow-right-square';
                linkCss = 'color: var(--primary-color);';
                tsCss = 'color: var(--primary-color); font-weight: bold;';
            }


            // create a timeline component for each sibling
            const timelineItem = html`
                <wiretap-timeline-item>
                    <span slot="time" style="${tsCss}">${tsFormat}</span>
                    <sl-icon name="${linkIcon}" slot="icon" style="${linkCss}"></sl-icon>
                    <div slot="content">${siblingComponent}</div>
                </wiretap-timeline-item>`

            timelineItems.push(timelineItem);

            tsDiff = transaction.httpResponse?.timestamp;
            //}
        });
        let tsTotalFormat = "";
        if (tsTotal > 1000) {
            tsTotalFormat = (tsTotal / 1000).toFixed(3) + " seconds";
        }
        if (tsTotal > 60000) {
            tsTotalFormat = ((tsTotal / 1000) / 60).toFixed(2) + " minutes";
        }
        if (tsTotal <= 1000) {
            tsTotalFormat = tsDiff.toFixed(3) + " ms";
        }

        this._siblings = siblings;
        return html`
            <div class="kv-overview">
                ${!hideKv ? paramKVComponent : null}
                <div class="empty-data chain-time">
                    <div class="request-chain-time-title">Chain Total Time</div>
                    <sl-icon name="stopwatch" class="chain-time-icon"></sl-icon>
                    <span class="total-time"><span class="time-value">${tsTotalFormat}</span></span>
                </div>
            </div>
            <wiretap-timeline>
                ${timelineItems.reverse()}
            </wiretap-timeline>
            <div>${siblings.length <= 0 ? this.noOtherLinks() : null}</div>`
    }

    noOtherLinks(): TemplateResult {
        return html`
            <div class="empty-data no-chain">
                <sl-icon name="link-45deg" class="binary-icon"></sl-icon>
                <br/>
                There are no other requests in this chain yet.
            </div>`
    }

    renderChainTabPanel(): TemplateResult {

        const selectChain = () => {
            if (this._siblings?.length > 0) {
                return html`
                    <div class="empty-data select-chain">
                        <sl-icon name="link-45deg" class="binary-icon"></sl-icon>
                        <br/>
                        Select a request from the chain.
                    </div>`
            }
            return null;
        }

        if (!this.hideChain) {
            return html`
                <sl-tab-panel name="chain">
                    <section class="chain-panel-divider">
                        <div class="chain-container">
                            ${this._currentLinks.map((linkMatch) =>
                                    html`${this.renderLinkMatch.bind(this)(linkMatch)}`
                            )}
                        </div>
                        <div class="chain-view-container">
                            ${this._chainTransactionView ? this._chainTransactionView : selectChain()}
                        </div>
                    </section>
                </sl-tab-panel>`
        }
        return null;
    }

    private syncLinks() {
        if (this._linkCache && this._httpTransaction) {
            const foundLinks = this._linkCache.findLinks(this._httpTransaction);
            if (foundLinks && foundLinks.length > 0) {

                // look at each link match and filter out any that contain siblings with the current transaction.
                // foundLinks.forEach((linkMatch) => {
                //     linkMatch.siblings = linkMatch.siblings.filter((sibling) => {
                //         return sibling.id !== this._httpTransaction.id;
                //     });
                // })

                this._currentLinks = foundLinks; // update state.
            } else {
                if (this._selectedTab === "chain") {
                    this._tabs.show("violations");
                }
                this._currentLinks = [];
            }
        }
    }
}