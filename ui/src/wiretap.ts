import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {CreateStoreManager, StoreManager} from "./ranch/store.manager";
import {HttpRequest, HttpResponse, HttpTransaction} from "./model/http_transaction";
import {Store} from "./ranch/store";
import {Bus, BusCallback, Channel, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";
import * as localforage from "localforage";

export const WiretapChannel = "wiretap-broadcast";
export const WiretapHttpTransactionStore = "http-transaction-store";
export const WiretapSelectedTransactionStore = "selected-transaction-store";
export const WiretapLocalStorage = "wiretap-local-storage";

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: StoreManager;
    private readonly _httpTransactionStore: Store<HttpTransaction>;
    private readonly _selectedTransactionStore: Store<HttpTransaction>;
    private readonly _bus: Bus;
    private readonly _wiretapChannel: Channel;
    private _channelSubscription: Subscription;

    constructor() {
        super();

        localforage.config({
            name: 'pb33f-wiretap',
            version: 1.0,
            storeName: 'wiretap_transactions',
        });

        // set up bus and stores
        this._bus = CreateBus();
        this._storeManager = CreateStoreManager();

        // create transaction store
        this._httpTransactionStore =
            this._storeManager.CreateStore<HttpTransaction>(WiretapHttpTransactionStore);

        // create selected transaction store
        this._selectedTransactionStore =
            this._storeManager.CreateStore<HttpTransaction>(WiretapSelectedTransactionStore);

        // set up wiretap channel
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);

        // load previous transactions from local storage.
        this.loadHistoryFromLocalStorage().then((previousTransactions: Map<string, HttpTransaction>) => {
            this._httpTransactionStore.populate(previousTransactions)
        });

        // handle incoming http transactions.
        this._channelSubscription = this._wiretapChannel.subscribe(this.wireHandler());

        // configure broker.
        const config = {
            brokerURL: 'ws://localhost:9091/ranch',
            heartbeatIncoming: 0,
            heartbeatOutgoing: 0,
        }
        // map and connect.
        this._bus.mapChannelToBrokerDestination("/topic/" + WiretapChannel, WiretapChannel)
        this._bus.connectToBroker(config)
    }


    async loadHistoryFromLocalStorage(): Promise<Map<string, HttpTransaction>> {
        return localforage.getItem<Map<string, HttpTransaction>>(WiretapLocalStorage);
    }


    wireHandler(): BusCallback {
        return (msg) => {
            const wiretapMessage = msg.payload as HttpTransaction

            const httpTransaction: HttpTransaction = {
                httpRequest: Object.assign(new HttpRequest(), wiretapMessage.httpRequest),
                id: wiretapMessage.id,
                requestValidation: wiretapMessage.requestValidation,
                responseValidation: wiretapMessage.responseValidation,
            }
            if (wiretapMessage.httpResponse) {
                httpTransaction.httpResponse = Object.assign(new HttpResponse(), wiretapMessage.httpResponse);
            }

            const existingTransaction: HttpTransaction = this._httpTransactionStore.get(httpTransaction.id)
            if (existingTransaction) {
                if (httpTransaction.httpResponse) {
                    existingTransaction.httpResponse = httpTransaction.httpResponse
                    existingTransaction.responseValidation = httpTransaction.responseValidation
                    this._httpTransactionStore.set(existingTransaction.id, existingTransaction)
                }
            } else {
                httpTransaction.timestamp = new Date().getTime();
                this._httpTransactionStore.set(httpTransaction.id, httpTransaction)
            }
        }
    }

    render() {
        const transaction =
            new HttpTransactionContainerComponent(this._httpTransactionStore, this._selectedTransactionStore)
        return html`${transaction}`
    }

}