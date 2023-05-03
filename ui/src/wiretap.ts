import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {CreateStoreManager, StoreManager} from "./ranch/store.manager";
import {HttpTransaction} from "./model/http_transaction";
import {Store} from "./ranch/store";
import {Bus, BusCallback, Channel, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";
import * as localforage from "localforage";

export const WiretapChannel = "wiretap-broadcast";
export const WiretapHttpTransactionStore = "http-transaction-store";
export const WiretapLocalStorage = "wiretap-local-storage";

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: StoreManager;
    private readonly _httpTransactionStore: Store<HttpTransaction>;
    private readonly _bus: Bus;
    private readonly _wiretapChannel: Channel;
    private _channelSubscription: Subscription;

    constructor() {
        super();

        localforage.config({
            name        : 'pb33f-wiretap',
            version     : 1.0,
            storeName   : 'wiretap_transactions',
        });

        // set up bus and stores
        this._bus = CreateBus();
        this._storeManager = CreateStoreManager();
        this._httpTransactionStore = this._storeManager.GetStore<HttpTransaction>(WiretapHttpTransactionStore);
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);

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
            const httpTransaction: HttpTransaction = msg.payload as HttpTransaction
            const existingTransaction: HttpTransaction = this._httpTransactionStore.get(httpTransaction.id)
            if (existingTransaction) {
                if (httpTransaction.httpResponse) {
                    existingTransaction.httpResponse = httpTransaction.httpResponse
                    this._httpTransactionStore.set(existingTransaction.id, existingTransaction)
                }
            } else {
                this._httpTransactionStore.set(httpTransaction.id, httpTransaction)
            }
        }
    }

    render() {
        const transaction = new HttpTransactionContainerComponent(this._httpTransactionStore)
        return html`${transaction}`
    }

}