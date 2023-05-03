import {customElement, property} from "lit/decorators.js";
import {html, LitElement, TemplateResult} from "lit";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import {HttpTransaction} from "@/model/http_transaction";
import transactionComponentCss from "@/components/transaction/transaction.component.css";
import Prism from 'prismjs'
import 'prismjs/components/prism-javascript' // Language
import 'prismjs/themes/prism-okaidia.css' // Theme


@customElement('http-transaction')
export class HttpTransactionComponent extends LitElement {

    static styles = transactionComponentCss
    private readonly _httpTransaction: HttpTransaction
    constructor(httpTransaction: HttpTransaction) {
        super();
        this._httpTransaction = httpTransaction
    }

    get transactionId(): string {
        return this._httpTransaction.id
    }


    render() {

        Prism.highlightAll();

        const req = this._httpTransaction.httpRequest;
        const resp = this._httpTransaction.httpResponse;
        let ok: TemplateResult
        if (this._httpTransaction.httpResponse) {
            ok = html`<pre>${resp.responseBody}</pre>`
        }

        const exchangeMethod = (method: string): string => {
            switch (method) {
                case 'GET':
                    return 'success'
                case 'POST':
                    return 'primary'
                case 'PUT':
                    return 'primary'
                case 'DELETE':
                    return 'danger'
                case 'PATCH':
                    return 'warning'
                case 'OPTIONS':
                    return 'neutral'
                case 'HEAD':
                    return 'neutral'
                case 'TRACE':
                    return 'neutral'
                default:
                    return 'primary'
            }
        }

         const highlight =Prism.highlight(resp.responseBody, Prism.languages['javascript'], 'javascript')


        const date = Date.now();

        return html`<section class="transaction">
            <header>
                <sl-tag variant="${exchangeMethod(req.method)}" class="method">${req.method}</sl-tag>
                ${req.url}
            </header>
            <main>
                
                
                ${date}
                
                <sl-tab-group>
                    <sl-tab slot="nav" panel="violations" class="tab">Violations</sl-tab>
                    <sl-tab slot="nav" panel="parameters" class="tab">Query Parameters</sl-tab>
                    <sl-tab slot="nav" panel="cookies" class="tab">Cookies</sl-tab>
                    <sl-tab slot="nav" panel="rheaders" class="tab">Headers</sl-tab>
                    <sl-tab slot="nav" panel="request-body" class="tab">Request Body</sl-tab>
                    <sl-tab slot="nav" 
                            panel="${this._httpTransaction.httpResponse ? 
                                    'response-body' : 'disabled'}" class="tab">Response Body</sl-tab>
                    
                    <sl-tab-panel name="violations">violations</sl-tab-panel>
                    <sl-tab-panel name="parameters">parameters.</sl-tab-panel>
                    <sl-tab-panel name="request-headers">request headers</sl-tab-panel>
                    <sl-tab-panel name="request-body">${req.requestBody}</sl-tab-panel>
                    <sl-tab-panel name="response-headers">response headers</sl-tab-panel>
                    <sl-tab-panel name="response-body">
                        <pre><code>${unsafeHTML(highlight)}</code></pre>  
                    </sl-tab-panel>
                </sl-tab-group>
                
                
                
            </main>
        </section>`
    }






}