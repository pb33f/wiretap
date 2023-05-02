import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {Store} from "@/ranch/store";
import {HttpTransaction} from '@/model/http_transaction';
import {HttpTransactionComponent} from "@/components/transaction/transaction.component";

@customElement('http-transaction-container')
export class HttpTransactionContainerComponent extends LitElement {
    private _httpTransactionStore: Store<HttpTransaction>;
    private _httpTransactions: Map<string, HttpTransactionContainer>

    constructor(store: Store<HttpTransaction>) {
        super()
        this._httpTransactionStore = store
        this._httpTransactions = new Map<string, HttpTransactionContainer>()
    }

    connectedCallback() {
        super.connectedCallback();
        this._httpTransactionStore.onAllChanges(this.handleTransactionChange.bind(this))
    }

    handleTransactionChange(key: string, value: HttpTransaction) {

        if (this._httpTransactions.has(value.id)) {
            const existingTransaction = this._httpTransactions.get(value.id)
            existingTransaction.Listener(value)
        } else {
            const container: HttpTransactionContainer = {
                Transaction: value,
                Listener: (trans: HttpTransaction) => {
                    this.requestUpdate();
                }
            }
            this._httpTransactions.set(value.id, container)
        }
    }


    render() {

        let transactions: HttpTransactionComponent[] = []
        this._httpTransactions.forEach(
            (v: HttpTransactionContainer) => {
                const comp = new HttpTransactionComponent(v.Transaction)
                transactions.push(comp)
            }
        )


        return html`${transactions}`
    }

}

interface HttpTransactionContainer {
    Transaction: HttpTransaction
    Listener: (update: HttpTransaction) => void
}