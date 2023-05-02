import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {CreateStoreManager, StoreManager} from "./ranch/store.manager";
import {HttpTransaction} from "./model/http_transaction";
import {Store} from "./ranch/store";
import {Bus, BusCallback, Channel, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";

export const WiretapChannel = "wiretap-broadcast";
export const WiretapHttpTransactionStore = "http-transaction-store";


@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private _storeManager: StoreManager;
    private _httpTransactionStore: Store<HttpTransaction>;
    private _bus: Bus;
    private _wiretapChannel: Channel;
    private _channelSubscription: Subscription;

    constructor() {
        super();

        // set up bus and stores
        this._bus = CreateBus();
        this._storeManager = CreateStoreManager();
        this._httpTransactionStore = this._storeManager.GetStore<HttpTransaction>(WiretapHttpTransactionStore);
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);

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