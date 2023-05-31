import {customElement, property, query} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpRequest, HttpResponse, HttpTransaction} from "./model/http_transaction";
import {Bag, BagManager, CreateBagManager} from "@pb33f/saddlebag";
import {Bus, BusCallback, Channel, CommandResponse, CreateBus, Subscription} from "@pb33f/ranch";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container.component";
import * as localforage from "localforage";
import {HeaderComponent} from "@/components/wiretap-header/header.component";
import {WiretapControls} from "@/model/controls";
import {
    GetCurrentSpecCommand,
    SpecChannel,
    WiretapChannel,
    WiretapControlsChannel, WiretapControlsKey, WiretapControlsStore,
    WiretapCurrentSpec,
    WiretapHttpTransactionStore,
    WiretapLocalStorage,
    WiretapSelectedTransactionStore,
    WiretapSpecStore
} from "@/model/constants";


@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: BagManager;
    private readonly _httpTransactionStore: Bag<HttpTransaction>;
    private readonly _selectedTransactionStore: Bag<HttpTransaction>;
    private readonly _controlsStore: Bag<WiretapControls>;
    private readonly _specStore: Bag<string>;

    private readonly _bus: Bus;
    private readonly _wiretapChannel: Channel;
    private readonly _specChannel: Channel;
    private readonly _wiretapControlsChannel: Channel;

    private _transactionChannelSubscription: Subscription;
    private _specChannelSubscription: Subscription;
    private _wiretapControlsSubscription: Subscription;

    private _transactionContainer: HttpTransactionContainerComponent;

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
        this._storeManager = CreateBagManager();

        // create transaction store
        this._httpTransactionStore =
            this._storeManager.createBag<HttpTransaction>(WiretapHttpTransactionStore);

        // create selected transaction store
        this._selectedTransactionStore =
            this._storeManager.createBag<HttpTransaction>(WiretapSelectedTransactionStore);

        // spec store
        this._specStore = this._storeManager.createBag<string>(WiretapSpecStore);

        // controls store
        this._controlsStore = this._storeManager.createBag<WiretapControls>(WiretapControlsStore);

        // set up wiretap channels
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);
        this._specChannel = this._bus.createChannel(SpecChannel);
        this._wiretapControlsChannel = this._bus.createChannel(WiretapControlsChannel);

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

        // handle incoming messages on different channels.
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
        this._bus.mapChannelToBrokerDestination("/queue/" + WiretapControlsChannel, WiretapControlsChannel)

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

            // get global delay
            const controls = this._controlsStore.get(WiretapControlsKey)
            if (controls.globalDelay > 0) {
                httpTransaction.delay = controls.globalDelay;
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
        if (this.violatedTransactions > 0) {
            this.complianceLevel = 100 - parseFloat(
                (this.violatedTransactions / (this.requestCount + this.responseCount) * 100)
                    .toFixed(2));
        } else {
            this.complianceLevel = 100;
        }
    }


    wipeData(e: CustomEvent) {
        const store = this._storeManager.getBag(e.detail)
        if (store) {
            store.reset();
        }
        this.responseCount = 0;
        this.requestCount = 0;
        this.violatedTransactions = 0;
        this.violationsCount = 0;
        this.calcComplianceLevel();
    }

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
                    @wipeData=${this.wipeData}
                    requests="${this.requestCount}"
                    responses="${this.responseCount}"
                    violations="${this.violationsCount}"
                    violationsDelta="${this.violatedTransactions}"
                    compliance="${this.complianceLevel}">
            </wiretap-header>
            ${transaction}`
    }

}