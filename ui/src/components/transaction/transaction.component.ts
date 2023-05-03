import {customElement, property} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpTransaction} from "@/model/http_transaction";


@customElement('http-transaction')
export class HttpTransactionComponent extends LitElement {
    private readonly _httpTransaction: HttpTransaction
    constructor(httpTransaction: HttpTransaction) {
        super();
        this._httpTransaction = httpTransaction
    }

    get transactionId(): string {
        return this._httpTransaction.id
    }

    render() {
        let ok = '';
        if (this._httpTransaction.httpResponse) {
            ok = "DONE"
        }
        return html`<div>
            <strong>ID: ${this._httpTransaction.id}</strong>
            ${ok}
        </div>`
    }



}