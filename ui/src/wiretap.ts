import {customElement, query, property} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {CreateStoreManager, StoreManager} from "./ranch/store.manager";
import {HttpRequest, HttpResponse, HttpTransaction} from "./model/http_transaction";
import {Store} from "./ranch/store";
import {Bus, BusCallback, Channel, CommandResponse, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";
import * as localforage from "localforage";
import {HeaderComponent} from "@/components/wiretap-header/header.component";
import {WiretapControls} from "@/model/controls";

export const WiretapChannel = "wiretap-broadcast";
export const SpecChannel = "specs";
export const WiretapHttpTransactionStore = "http-transaction-store";
export const WiretapSelectedTransactionStore = "selected-transaction-store";
export const WiretapSpecStore = "wiretap-spec-store";
export const WiretapControlsStore = "wiretap-controls-store";
export const WiretapCurrentSpec = "current-spec";
export const GetCurrentSpecCommand = "get-current-spec";

export const ChangeDelayCommand = "change-delay-request";
export const WiretapLocalStorage = "wiretap-transactions";

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: StoreManager;
    private readonly _httpTransactionStore: Store<HttpTransaction>;
    private readonly _selectedTransactionStore: Store<HttpTransaction>;
    private readonly _controlsStore: Store<WiretapControls>;
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
    violatedTransactions = 0.0;

    @property({type: Number})
    complianceLevel: number = 100.0;


    constructor() {
        super();

        //configure local storage
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

        // spec store
        this._specStore = this._storeManager.CreateStore<string>(WiretapSpecStore);

        // controls store
        this._controlsStore = this._storeManager.CreateStore<WiretapControls>(WiretapSpecStore);

        // set up wiretap channels
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);
        this._specChannel = this._bus.createChannel(SpecChannel);

        // load previous transactions from local storage.
        this.loadHistoryFromLocalStorage().then((previousTransactions: Map<string, HttpTransaction>) => {
            this._httpTransactionStore.populate(previousTransactions)

            // calculate counts from stored state.
            this.calculateMetricsFromState(previousTransactions);
        });

        this.loadSpecFromLocalStorage().then((spec: string) => {
            if (!spec || spec.length <= 0) {
                // nothing in local storage, request from server.
                this.requestSpec()
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

        this._bus.connectToBroker(config);
    }

    calculateMetricsFromState(previousTransactions: Map<string, HttpTransaction>) {
        let requests = 0;
        let responses = 0;
        let violations = 0;
        let violated = 0.0
        previousTransactions.forEach((transaction: HttpTransaction) => {
            requests++;
            if (transaction.httpResponse) {
                responses++;
            }
            if (transaction.requestValidation) {
                violated += 0.5;
                violations += transaction.requestValidation.length
            }
            if (transaction.responseValidation) {
                violated += 0.5;
                violations += transaction.responseValidation.length;
            }
        });
        this.requestCount = requests;
        this.responseCount = responses;
        this.violationsCount = violations;
        this.violatedTransactions = violated;
        this.calcComplianceLevel();


    }

    requestSpec() {
        this._bus.getClient().publish({
            destination: "/pub/queue/specs",
            body: JSON.stringify({requestCommand: GetCurrentSpecCommand}),
        })
    }



    async loadSpecFromLocalStorage(): Promise<string> {
        return localforage.getItem<string>(WiretapCurrentSpec);
    }

    async loadHistoryFromLocalStorage(): Promise<Map<string, HttpTransaction>> {
        return localforage.getItem<Map<string, HttpTransaction>>(WiretapLocalStorage);
    }

    specHandler(): BusCallback<CommandResponse> {
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

            if (wiretapMessage.requestValidation && wiretapMessage.requestValidation.length > 0) {
                this.violatedTransactions += 0.5;
                this.violationsCount += wiretapMessage.requestValidation.length
            }

            if (wiretapMessage.httpResponse) {
                this.responseCount++;
                httpTransaction.httpResponse = Object.assign(new HttpResponse(), wiretapMessage.httpResponse);
                if (wiretapMessage.responseValidation && wiretapMessage.responseValidation.length > 0) {
                    this.violatedTransactions += 0.5;
                    this.violationsCount += wiretapMessage.responseValidation.length
                }
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

            this.calcComplianceLevel();
        }
    }

    calcComplianceLevel(): void {
        this.complianceLevel = 100 - parseFloat((this.violatedTransactions / (this.requestCount + this.responseCount)*100).toFixed(2));
    }

    private _transactionContainer: HttpTransactionContainerComponent;

    render() {

        let transaction: HttpTransactionContainerComponent
        if (this._transactionContainer) {
            transaction = this._transactionContainer;
        } else {
            transaction = new HttpTransactionContainerComponent(
                this._httpTransactionStore,
                this._selectedTransactionStore,
                this._specStore);
            this._transactionContainer = transaction;
        }
        return html`
            <wiretap-header
                    requests="${this.requestCount}"
                    responses="${this.responseCount}"
                    violations="${this.violationsCount}"
                    violationsDelta="${this.violatedTransactions}"
                    compliance="${this.complianceLevel}">
            </wiretap-header>
            ${transaction}`
    }

}