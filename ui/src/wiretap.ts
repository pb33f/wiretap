import {customElement, query, property} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {CreateStoreManager, StoreManager} from "./ranch/store.manager";
import {HttpRequest, HttpResponse, HttpTransaction} from "./model/http_transaction";
import {Store} from "./ranch/store";
import {Bus, BusCallback, Channel, CommandResponse, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";
import * as localforage from "localforage";
import {HeaderComponent} from "@/components/wiretap-header/header.component";


export const WiretapChannel = "wiretap-broadcast";
export const SpecChannel = "specs";

export const WiretapHttpTransactionStore = "http-transaction-store";
export const WiretapSelectedTransactionStore = "selected-transaction-store";

export const WiretapSpecStore = "wiretap-spec-store";
export const WiretapCurrentSpec = "current-spec";
export const GetCurrentSpecCommand = "get-current-spec";



export const WiretapLocalStorage = "wiretap-transactions";

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: StoreManager;
    private readonly _httpTransactionStore: Store<HttpTransaction>;
    private readonly _selectedTransactionStore: Store<HttpTransaction>;
    private readonly _specStore: Store<string>;

    private readonly _bus: Bus;
    private readonly _wiretapChannel: Channel;
    private readonly _specChannel: Channel;
    private _transactionChannelSubscription: Subscription;
    private _specChannelSubscription: Subscription;

    @query("wiretap-header")
    private _wiretapHeader: HeaderComponent;

    @property({type: Number})
    requestCount = 0;
    @property({type: Number})
    responseCount = 0;
    @property({type: Number})
    violationsCount = 0;
    @property({type: Number})
    complianceLevel: number = 0;

    constructor() {
        super();

        localforage.config({
            name: 'pb33f-wiretap',
            version: 1.0,
            storeName: 'wiretap',
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

        this._specStore = this._storeManager.CreateStore<string>(WiretapSpecStore);

        // set up wiretap channel
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);
        this._specChannel = this._bus.createChannel(SpecChannel);

        // load previous transactions from local storage.
        this.loadHistoryFromLocalStorage().then((previousTransactions: Map<string, HttpTransaction>) => {
            this._httpTransactionStore.populate(previousTransactions)
        });

        this.loadSpecFromLocalStorage().then((spec: string) => {
            if (!spec || spec.length <= 0) {
                this._bus.getClient().publish({
                    destination: "/pub/queue/specs",
                    body: JSON.stringify({requestCommand: GetCurrentSpecCommand}),
                })
            } else {
                this._specStore.set(WiretapCurrentSpec, spec);
            }
        });

        // handle incoming http transactions.
        this._transactionChannelSubscription = this._wiretapChannel.subscribe(this.wireTransactionHandler());
        this._specChannelSubscription = this._specChannel.subscribe(this.specHandler());

        // configure broker.
        const config = {
            brokerURL: 'ws://localhost:9090/ranch',
            heartbeatIncoming: 0,
            heartbeatOutgoing: 0,
        }
        // map and connect.
        this._bus.mapChannelToBrokerDestination("/topic/" + WiretapChannel, WiretapChannel)
        this._bus.mapChannelToBrokerDestination("/queue/" + SpecChannel, SpecChannel)

        //setTimeout(() => {
            this._bus.connectToBroker(config)
        //},4000);
;
    }


    async loadSpecFromLocalStorage(): Promise<string> {
        return localforage.getItem<string>(WiretapCurrentSpec);
    }

    async loadHistoryFromLocalStorage(): Promise<Map<string, HttpTransaction>> {
        return localforage.getItem<Map<string, HttpTransaction>>(WiretapLocalStorage);
    }

    specHandler(): BusCallback<CommandResponse>{
        return (msg) => {
            const decoded = atob(msg.payload.payload);
            this._specStore.set(WiretapCurrentSpec, decoded)
            localforage.setItem(WiretapCurrentSpec, decoded);
        }
    }



    wireTransactionHandler(): BusCallback {
        return (msg) => {
            const wiretapMessage = msg.payload as HttpTransaction
            const httpTransaction: HttpTransaction = {
                httpRequest: Object.assign(new HttpRequest(), wiretapMessage.httpRequest),
                id: wiretapMessage.id,
                requestValidation: wiretapMessage.requestValidation,
                responseValidation: wiretapMessage.responseValidation,
            }
            if (wiretapMessage.httpResponse) {
                this.responseCount++;
                console.log("ho ho", this.responseCount);
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
                this.requestCount++;
                httpTransaction.timestamp = new Date().getTime();
                this._httpTransactionStore.set(httpTransaction.id, httpTransaction)
            }
            console.log("hey hey", this.requestCount);
        }
    }

    render() {

        // TODO: re-work this to allow the header state to update without needing a rebuild of the transaction container.

        const transaction =
            new HttpTransactionContainerComponent(
                this._httpTransactionStore,
                this._selectedTransactionStore,
                this._specStore);

        return html`<wiretap-header
            requests="${this.requestCount}"
            responses="${this.responseCount}"
            violations="${this.violationsCount}"
            compliance="${this.complianceLevel}">
        </wiretap-header>
        ${transaction}`
    }

}