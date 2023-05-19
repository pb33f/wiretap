import {customElement, state} from "lit/decorators.js";
import {html} from "lit";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import {map} from "lit/directives/map.js";
import {LitElement, TemplateResult} from "lit";

import {HttpTransaction} from "@/model/http_transaction";
import transactionViewComponentCss from "./transaction-view.component.css";
import {KVViewComponent} from "@/components/kv-view/kv-view.component";

import prismCss from "@/components/prism.css";
import Prism from 'prismjs';
import 'prismjs/components/prism-javascript';
import 'prismjs/themes/prism-okaidia.css';

@customElement('http-transaction-view')
export class HttpTransactionViewComponent extends LitElement {

    static styles = [prismCss, transactionViewComponentCss];

    @state()
    private _httpTransaction: HttpTransaction

    private readonly _requestHeadersView: KVViewComponent;
    private readonly _responseHeadersView: KVViewComponent;
    private readonly _requestCookiesView: KVViewComponent;
    private readonly _responseCookiesView: KVViewComponent;


    constructor() {
        super();
        this._requestHeadersView = new KVViewComponent();
        this._requestCookiesView = new KVViewComponent();
        this._responseHeadersView = new KVViewComponent();
        this._responseCookiesView = new KVViewComponent();
    }

    set httpTransaction(value: HttpTransaction) {
        this._httpTransaction = value;
        if (this._requestHeadersView && value.httpRequest) {
            this._requestHeadersView.data = value.httpRequest.extractHeaders();
            this._requestCookiesView.data = value.httpRequest.extractCookies();
        }
        if (this._responseHeadersView && value.httpResponse) {
            this._responseHeadersView.data = value.httpResponse.extractHeaders();
            this._responseCookiesView.data = value.httpResponse.extractCookies();
        }
    }

    render() {

        console.log(this._httpTransaction);

        if (this._httpTransaction) {

            const req = this._httpTransaction?.httpRequest;
            const resp = this._httpTransaction?.httpResponse;

            const requestViolations: TemplateResult = html`
                ${this._httpTransaction?.requestValidation?.length > 0 ? html`<h3>Request Violations</h3>` : html``}
                ${map(this._httpTransaction.requestValidation, (i) => {
                    return html`
                        <sl-details summary="${i.message}">
                            ${i.reason}
                        </sl-details>
                    `})}`;

            const responseViolations: TemplateResult = html`
                ${this._httpTransaction?.responseValidation?.length > 0 ? html`<h3>Response Violations</h3>` : html``}
                ${map(this._httpTransaction.responseValidation, (i) => {
                    return html`
                        <sl-details summary="${i.message}">
                            ${i.reason}
                        </sl-details>
                    `})}`;


            let highlight: string;
            if (resp && resp.responseBody) {
                highlight = Prism.highlight(resp.responseBody, Prism.languages['javascript'], 'javascript')
            }

            const tabGroup: TemplateResult = html`
                <sl-tab-group>
                    <sl-tab slot="nav" panel="violations" class="tab">Violations</sl-tab>
                    <sl-tab slot="nav" panel="parameters" class="tab">Query Parameters</sl-tab>
                    <sl-tab slot="nav" panel="request" class="tab">Request</sl-tab>
                    <sl-tab slot="nav" panel="response" class="tab">Response</sl-tab>
                    <sl-tab-panel name="violations">
                        ${requestViolations}
                        ${responseViolations}
                    </sl-tab-panel>
                    <sl-tab-panel name="parameters">
                        parameters.
                    </sl-tab-panel>
                    <sl-tab-panel name="request">
                        <sl-tab-group class="secondary-tabs" placement="start">
                            <sl-tab slot="nav" panel="request-headers" class="tab-secondary">Headers</sl-tab>
                            <sl-tab slot="nav" panel="request-cookies" class="tab-secondary">Cookies</sl-tab>
                            <sl-tab slot="nav" panel="request-body" class="tab-secondary">Body</sl-tab>
                            <sl-tab-panel name="request-headers">
                                ${this._requestHeadersView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-cookies">
                                ${this._requestCookiesView}
                            </sl-tab-panel>
                            <sl-tab-panel name="request-body">
                                ${req.requestBody}
                            </sl-tab-panel>
                        </sl-tab-group>
                    </sl-tab-panel>
                    <sl-tab-panel name="response">
                        <sl-tab-group class="secondary-tabs" placement="start">
                            <sl-tab slot="nav" panel="response-headers" class="tab-secondary">Headers</sl-tab>
                            <sl-tab slot="nav" panel="response-cookies" class="tab-secondary">Cookies</sl-tab>
                            <sl-tab slot="nav" panel="response-body" class="tab-secondary">Body</sl-tab>
                            <sl-tab-panel name="response-headers">
                                ${this._responseHeadersView}
                            </sl-tab-panel>
                            <sl-tab-panel name="response-cookies">
                                ${this._responseCookiesView}
                            </sl-tab-panel>
                            <sl-tab-panel name="response-body">
                                <pre><code>${unsafeHTML(highlight)}</code></pre>
                            </sl-tab-panel>
                        </sl-tab-group>
                    </sl-tab-panel>
                </sl-tab-group>
            `


            return html`${tabGroup}`
        } else {
            return html`select a transaction`
        }


    }


}