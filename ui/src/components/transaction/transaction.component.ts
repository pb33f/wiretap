import {customElement} from "lit/decorators.js";
import {LitElement} from "lit";
import {HttpTransaction} from "@/model/http_transaction";


@customElement('http-transaction')
export class HttpTransactionComponent extends LitElement {
    private _httpTransaction: HttpTransaction
    constructor(httpTransaction: HttpTransaction) {
        super();
        this._httpTransaction = httpTransaction
    }

    render() {
        return `<div><strong>ID: ${this._httpTransaction.id}</strong>`
    }



}