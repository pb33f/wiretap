import {customElement, state} from "lit/decorators.js";
import {html, LitElement, TemplateResult} from "lit";
import {HttpTransaction} from "@/model/http_transaction";
import transactionComponentCss from "@/components/transaction/transaction-item.css";
import {ExchangeMethod} from "@/model/exchange_method";
import Prism from 'prismjs'
import 'prismjs/components/prism-javascript'
import 'prismjs/themes/prism-okaidia.css'
import {HttpTransactionSelectedEvent} from "@/model/events";
import sharedCss from "@/components/shared.css";
import {Filter, WiretapFilters} from "@/model/controls";


@customElement('http-transaction-item')
export class HttpTransactionItemComponent extends LitElement {

    static styles = [sharedCss, transactionComponentCss]

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

        let delay: TemplateResult;

        if (this._httpTransaction.delay > 0) {
            delay = html`<div class="delay"><sl-icon name="hourglass-split" ></sl-icon>${this._httpTransaction.delay}ms</div>`
        }

        return html`
            <div class="${tClass}" @click="${this.setActive}">
                <header>
                    <sl-tag variant="${ExchangeMethod(req.method)}" class="method">${req.method}</sl-tag>
                    ${decodeURI(req.url)}
                </header>
               ${delay}
                <div class="transaction-status">
                    ${statusIcon}
                </div>
            </div>`
    }
}