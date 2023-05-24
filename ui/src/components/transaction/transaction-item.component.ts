import {customElement, state} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpTransaction} from "@/model/http_transaction";
import transactionComponentCss from "@/components/transaction/transaction-item.component.css";
import Prism from 'prismjs'
import 'prismjs/components/prism-javascript' // Language
import 'prismjs/themes/prism-okaidia.css'
import {HttpTransactionSelectedEvent} from "@/model/events"; // Theme

@customElement('http-transaction-item')
export class HttpTransactionItemComponent extends LitElement {

    static styles = transactionComponentCss

    @state()
    _httpTransaction: HttpTransaction

    @state()
    _active = false;

    private _processing = false;

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
        this._processing = req && !resp;

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
                default:
                    return 'neutral'
            }
        }

        let tClass = "transaction";
        if (this._active) {
            tClass += " active";
        }

        let statusIcon = html``

        if (this._processing) {
            statusIcon = html`<sl-spinner class="spinner"></sl-spinner>`
        } else {
            if (req && resp) {
                if (this._httpTransaction.requestValidation?.length > 0 ||
                    this._httpTransaction.responseValidation?.length > 0) {
                    statusIcon = html`<sl-icon name="exclamation-circle" class="invalid"></sl-icon>`
                } else {
                    statusIcon = html`<sl-icon name="check-lg" class="valid"></sl-icon>`
                }
            }
        }

        return html`
            <div class="${tClass}" @click="${this.setActive}">
                <header>
                    <sl-tag variant="${exchangeMethod(req.method)}" class="method">${req.method}</sl-tag>
                    ${decodeURI(req.url)}
                </header>
                <div class="transaction-status">
                    ${statusIcon}
                </div>
            </div>`
    }


}