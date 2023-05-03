import {customElement, state} from "lit/decorators.js";
import {repeat} from 'lit/directives/repeat.js';
import {html, LitElement} from "lit";
import {Store} from "@/ranch/store";
import {HttpTransaction} from '@/model/http_transaction';
import {HttpTransactionComponent} from "./transaction.component";
import localforage from "localforage";
import {WiretapLocalStorage} from "@/wiretap";
import transactionContainerComponentCss from "./transaction-container.component.css";

@customElement('http-transaction-container')
export class HttpTransactionContainerComponent extends LitElement {

    static styles = transactionContainerComponentCss;

    private _httpTransactionStore: Store<HttpTransaction>;

    @state()
    _httpTransactions: Map<string, HttpTransactionContainer>

    constructor(store: Store<HttpTransaction>) {
        super()
        this._httpTransactionStore = store
        this._httpTransactions = new Map<string, HttpTransactionContainer>()
    }

    connectedCallback() {
        super.connectedCallback();
        this._httpTransactionStore.onAllChanges(this.handleTransactionChange.bind(this))
        this._httpTransactionStore.onPopulated((storeData: Map<string, HttpTransaction>) => {
            // rebuild our internal state
            const savedTransactions: Map<string, HttpTransactionContainer> = new Map<string, HttpTransactionContainer>()
            storeData.forEach((value: HttpTransaction, key: string) => {
                const container: HttpTransactionContainer = {
                    Transaction: value,
                    Listener: (update: HttpTransaction) => {
                        this.requestUpdate();
                    }
                }
                savedTransactions.set(key, container)
            });
            // save our internal state.
            this._httpTransactions = savedTransactions
        });
    }

    handleTransactionChange(key: string, value: HttpTransaction) {
        if (this._httpTransactions.has(value.id)) {
            const existingTransaction = this._httpTransactions.get(value.id)
            existingTransaction.Listener(value)
        } else {
            const container: HttpTransactionContainer = {
                Transaction: value,
                Listener: (trans: HttpTransaction) => {
                    // update db.
                    let exp = this._httpTransactionStore.export()
                    localforage.setItem<Map<string, HttpTransaction>>
                    (WiretapLocalStorage, exp).then(
                        () => {
                            this.requestUpdate();
                        }
                    ).catch(
                        (err) => {
                            console.error(err)
                        }
                    )
                }
            }
            this._httpTransactions.set(value.id, container)
            this.requestUpdate();
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

        return html`<section class="transactions">
            ${repeat(transactions.reverse(),
            (t: HttpTransactionComponent) => t.transactionId,
            (t: HttpTransactionComponent) => t)}
        </section>`
    }

}

interface HttpTransactionContainer {
    Transaction: HttpTransaction
    Listener: (update: HttpTransaction) => void
}