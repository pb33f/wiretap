import {customElement, property, state} from "lit/decorators.js";
import {html, LitElement, TemplateResult} from "lit";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import {HttpTransaction} from "@/model/http_transaction";
import transactionComponentCss from "@/components/transaction/transaction-item.component.css";
import Prism from 'prismjs'
import 'prismjs/components/prism-javascript' // Language
import 'prismjs/themes/prism-okaidia.css' // Theme

export const HttpTransactionSelectedEvent = "httpTransactionSelected";

@customElement('http-transaction-item')
export class HttpTransactionItemComponent extends LitElement {

    static styles = transactionComponentCss

    @state()
    _httpTransaction: HttpTransaction

    @state()
    _active = false;

    constructor(httpTransaction: HttpTransaction) {
        super();
        this._httpTransaction = httpTransaction
    }

    get transactionId(): string {
        return this._httpTransaction.id
    }

    get active(): boolean {
        return this._active;
    }

    set httpTransaction(value: HttpTransaction) {
        this._httpTransaction = value;
    }

    get httpTransaction(): HttpTransaction {
        return this._httpTransaction;
    }

    disable() {
        this._active = false;
    }

    setActive(): void {
        if (this._active) {
            return
        }
        this._active = true;
        const options = {
            detail: this._httpTransaction,
            bubbles: true,
            composed: true,
        };
        this.dispatchEvent(
            new CustomEvent<HttpTransaction>(HttpTransactionSelectedEvent, options)
        );
    }


    render() {

        Prism.highlightAll();

        const req = this._httpTransaction?.httpRequest;
        const resp = this._httpTransaction?.httpResponse;


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



        let tClass = "transaction";
        if (this._active) {
            tClass += " active";
        }

        return html`
            <div class="${tClass}" @click="${this.setActive}">
                <header>
                    <sl-tag variant="${exchangeMethod(req.method)}" class="method">${req.method}</sl-tag>
                    ${req.url} ${tClass}
                </header>
            </div>`
    }


}