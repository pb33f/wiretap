import {customElement, state} from "lit/decorators.js";
import {LitElement, TemplateResult} from "lit";
import {HttpTransaction} from "@/model/http_transaction";
import {html} from "lit";

import Prism from 'prismjs';
import 'prismjs/components/prism-javascript';
import 'prismjs/themes/prism-okaidia.css';
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import transactionViewComponentCss from "@/components/transaction/transaction-view.component.css";

@customElement('http-transaction-view')
export class HttpTransactionViewComponent extends LitElement {

    static styles = transactionViewComponentCss

    @state()
    _httpTransaction: HttpTransaction

    constructor() {
        super();
    }

    set httpTransaction(value: HttpTransaction) {
        this._httpTransaction = value;
    }

    render() {


        if (this._httpTransaction) {

            const req = this._httpTransaction?.httpRequest;
            const resp = this._httpTransaction?.httpResponse;

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
              

                <sl-tab-panel name="violations">violations</sl-tab-panel>
                <sl-tab-panel name="parameters">
                    parameters.
                </sl-tab-panel>
                <sl-tab-panel name="request">

                    <sl-tab-group class="secondary-tabs">
                        <sl-tab slot="nav" panel="request-headers" class="tab-secondary">Headers</sl-tab>
                        <sl-tab slot="nav" panel="request-cookies" class="tab-secondary">Cookies</sl-tab>
                        <sl-tab slot="nav" panel="request-body" class="tab-secondary">Body</sl-tab>
                        <sl-tab-panel name="request-headers">
                            mooo.
                        </sl-tab-panel>
                        <sl-tab-panel name="request-cookies">
                            mooo.
                        </sl-tab-panel>
                        <sl-tab-panel name="request-body">
                            mooo.
                        </sl-tab-panel>
                    </sl-tab-group>
                </sl-tab-panel>
                <sl-tab-panel name="response">
                    
                    
                    ${req.requestBody}

                    <pre><code>${unsafeHTML(highlight)}</code></pre>
                </sl-tab-panel>
            </sl-tab-group>
        `


            return html`${tabGroup}`
        } else {
            return html`hheeeeeeyyyyyyyyyyy`
        }


    }


}