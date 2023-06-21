import {customElement, property, query, state} from "lit/decorators.js";
import {html} from "lit";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import {map} from "lit/directives/map.js";
import {LitElement, TemplateResult} from "lit";

import {BuildLiveTransactionFromState, HttpRequest, HttpResponse, HttpTransaction} from "@/model/http_transaction";
import transactionViewComponentCss from "./transaction-view.css";
import {KVViewComponent} from "@/components/kv-view/kv-view";

import prismCss from "@/components/prism.css";
import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-xml-doc';
import 'prismjs/themes/prism-okaidia.css';
import sharedCss from "@/components/shared.css";
import {SlTab, SlTabGroup} from "@shoelace-style/shoelace";
import {
    ContentTypeFormEncoded,
    ContentTypeHtml,
    ContentTypeJSON,
    ContentTypeMultipartForm,
    ContentTypeOctetStream,
    ContentTypeXML,
    ExtractContentTypeFromRequest,
    ExtractContentTypeFromResponse,
    IsHtmlContentType,
    IsOctectStreamContentType,
    IsXmlContentType, FormPart
} from "@/model/extract_content_type";
import {ExtractHTTPCodeDefinition, ExtractStatusStyleFromCode} from "@/model/extract_status";
import {Property, PropertyViewComponent} from "@/components/property-view/property-view";
import {LinkMatch, TransactionLinkCache} from "@/model/link_cache";
import {HttpTransactionItemComponent} from "@/components/transaction/transaction-item";
import {HttpTransactionSelectedEvent} from "@/model/events";
import {Bag, GetBagManager} from "@pb33f/saddlebag";
import {WiretapHttpTransactionStore} from "@/model/constants";

@customElement('http-transaction-view')
export class HttpTransactionViewComponent extends LitElement {

    static styles = [prismCss, sharedCss, transactionViewComponentCss];

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
    private _linkCache: TransactionLinkCache;

    // into the matrix.
    private _chainTransactionView: HttpTransactionViewComponent;
    private readonly _httpTransactionStore: Bag<HttpTransaction>;

    @state()
    private _currentLinks: LinkMatch[];

    constructor() {
        super();
        this._requestHeadersView = new KVViewComponent();
        this._requestCookiesView = new KVViewComponent();
        this._requestCookiesView.keyLabel = 'Cookie Name';
        this._responseHeadersView = new KVViewComponent();
        this._responseCookiesView = new KVViewComponent();
        this._responseCookiesView.keyLabel = 'Cookie Name';
        this._requestQueryView = new KVViewComponent();
        this._requestQueryView.keyLabel = 'Query Key';
        this._httpTransactionStore =
            GetBagManager().getBag<HttpTransaction>(WiretapHttpTransactionStore);
    }

    set linkCache(value: TransactionLinkCache) {
        this._linkCache = value;
        console.log('fappppppo', this._linkCache);
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
    }

    tabSelected(event: CustomEvent) {
        console.log('tab selected');
        this._selectedTab = event.detail.name;
        //this.syncLinks();
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

            const tabGroup: TemplateResult = html`
                <sl-tab-group id="tabs" @sl-tab-show=${this.tabSelected}>
                    <sl-tab slot="nav" panel="violations" id="violation-tab" class="tab">${violations}</sl-tab>
                    <sl-tab slot="nav" panel="request" class="tab">Request</sl-tab>
                    <sl-tab slot="nav" panel="response" class="tab">Response</sl-tab>
                    ${this._currentLinks?.length > 0 ? html`<sl-tab slot="nav" panel="chain" class="tab">Chain</sl-tab>`: null}
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
                            <sl-tab slot="nav" panel="request-body" class="tab-secondary">Body</sl-tab>
                            <sl-tab slot="nav" panel="request-query" class="tab">Query</sl-tab>
                            <sl-tab slot="nav" panel="request-headers" class="tab-secondary">Headers</sl-tab>
                            <sl-tab slot="nav" panel="request-cookies" class="tab-secondary">Cookies</sl-tab>
                            <sl-tab-panel name="request-headers">
                                ${this._requestHeadersView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-cookies">
                                ${this._requestCookiesView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-body">
                                ${this.renderRequestBody(req)}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-query">
                                ${this._requestQueryView}
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
                                ${this.renderResponseBody(resp)}
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
        console.log('a nice meal', event.detail);
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
        linkMatch.siblings.forEach((sibling) => {
            const transaction = BuildLiveTransactionFromState(this._httpTransactionStore.get(sibling.id));
            if (transaction.id !== this._httpTransaction.id) {
                const siblingComponent = new HttpTransactionItemComponent(transaction, this._linkCache);
                siblingComponent.hideControls = true;
                if (this._chainTransactionView && this._chainTransactionView.httpTransaction?.id == transaction.id) {
                    siblingComponent.setActive();
                } else {
                    siblingComponent.disable();
                }
                siblingComponent.addEventListener(HttpTransactionSelectedEvent, this.chainTransactionSelected.bind(this));
                siblings.push(siblingComponent);
            }
        })
        return html`${!hideKv? paramKVComponent: null }${siblings}${siblings.length <= 0 ? 'no more requests in chain...': null}`
    }


    renderChainTabPanel(): TemplateResult {
        if (!this.hideChain) {

            let selectOrBreadcrumb: TemplateResult;
            //if (!this._chainTransactionView) {
                selectOrBreadcrumb = html`${this._currentLinks.map((linkMatch) =>
                            html`${this.renderLinkMatch.bind(this)(linkMatch)}`
                    )}`
            // } else {
            //     selectOrBreadcrumb = html`NO`
            // }


            return html`
                <sl-tab-panel name="chain">
                    ${selectOrBreadcrumb}
                    ${this._chainTransactionView}

                </sl-tab-panel>`
        }
        return null;
    }

    parseFormEncodedData(data: string): Map<string, string> {
        const map = new Map<string, string>();
        const pairs = data.split('&');
        for (const pair of pairs) {
            const [key, value] = pair.split('=');
            map.set(decodeURI(key), decodeURI(value));
        }
        return map;
    }

    private syncLinks() {
        if (this._linkCache && this._httpTransaction) {
            const foundLinks = this._linkCache.findLinks(this._httpTransaction);
            if (foundLinks && foundLinks.length > 0) {

                // look at each link match and filter out any that contain siblings with the current transaction.
                foundLinks.forEach((linkMatch) => {
                    linkMatch.siblings = linkMatch.siblings.filter((sibling) => {
                        return sibling.id !== this._httpTransaction.id;
                    });
                })

                this._currentLinks = foundLinks; // update state.
            } else {
                if (this._selectedTab === "chain") {
                    this._tabs.show("violations");
                }
                this._currentLinks = [];
            }
        }
    }

    renderRequestBody(req: HttpRequest): TemplateResult {

        const exct = ExtractContentTypeFromRequest(req)
        const ct = html` <span class="contentType">
            Content Type: <strong>${exct}</strong>
        </span>`;

        switch (exct) {
            case ContentTypeJSON:
                return html`${ct}
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(req.requestBody), null, 2),
                            Prism.languages.json, 'json'))}</code></pre>`;

            case ContentTypeXML:
                return html`${ct}
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(req.requestBody), null, 2),
                            Prism.languages.xml, 'xml'))}</code></pre>`;

            case ContentTypeOctetStream:
                return html`${ct}
                    <div class="empty-data">
                        <sl-icon name="file-binary" class="binary-icon"></sl-icon>
                        <br/>
                        [ binary data will not be rendered ]
                    </div>`;
            case ContentTypeHtml:
                return html`${ct}
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(req.requestBody), null, 2),
                            Prism.languages.xml, 'xml'))}</code></pre>`;

            case ContentTypeFormEncoded:
                const kv = new KVViewComponent();
                kv.keyLabel = "Form Key";
                kv.data = this.parseFormEncodedData(req.requestBody);
                return html`${ct}${kv}`

            case ContentTypeMultipartForm:
                const formProps = new PropertyViewComponent()
                formProps.propertyLabel = "Form Key";
                formProps.typeLabel = "Type";

                // extract pre-rendered form data from wiretap
                const parts: FormPart[] = JSON.parse(req.requestBody) as FormPart[];
                for (const part of parts) {
                    if (part.value?.length > 0) {
                        part.type = 'field';
                    }
                    if (part.files?.length > 0) {
                        part.type = 'file';
                    }
                }

                formProps.data = parts;

                return html`${ct}${formProps}`

            default:
                return html`${ct}
                <pre>${req.requestBody}</pre>`
        }

    }

    renderResponseBody(resp: HttpResponse): TemplateResult {

        const exct = ExtractContentTypeFromResponse(resp)
        const ct = html` <span class="contentType">
            Content Type: <strong>${exct}</strong>
        </span>`;

        switch (exct) {
            case ContentTypeJSON:
                return html`${ct}
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(resp.responseBody), null, 2),
                            Prism.languages.json, 'json'))}</code></pre>`;
            case ContentTypeXML:
                return html`
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(resp.responseBody), null, 2),
                            Prism.languages.xml, 'xml'))}</code></pre>`;
            case ContentTypeOctetStream:
                return html`${ct}
                    <div class="empty-data">
                        <sl-icon name="file-binary" class="binary-icon"></sl-icon>
                        <br/>
                        [ binary data will not be rendered ]
                    </div>`;
            case ContentTypeHtml:
                return html`${ct}
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(resp.responseBody), null, 2),
                            Prism.languages.xml, 'xml'))}</code></pre>`;

            default:
                return html`${ct}
                    <pre>${resp.responseBody}</pre>`
        }

    }


}