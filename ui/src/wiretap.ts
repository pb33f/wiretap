import {customElement, property, query} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpRequest, HttpResponse, HttpTransaction, HttpTransactionBase} from "./model/http_transaction";
import {Bag, BagManager, CreateBagManager} from "@pb33f/saddlebag";
import {Bus, BusCallback, Channel, CommandResponse, CreateBus, Subscription} from "@pb33f/ranch";
import {HttpTransactionContainerComponent} from "./components/transaction/transaction-container";
import * as localforage from "localforage";
import {HeaderComponent} from "@/components/wiretap-header/header";
import {WiretapControls, WiretapFilters} from "@/model/controls";
import {
    GetCurrentSpecCommand, NoSpec, QueuePrefix,
    SpecChannel, TopicPrefix,
    WiretapChannel, WiretapConfigurationChannel,
    WiretapControlsChannel, WiretapControlsKey, WiretapControlsStore,
    WiretapCurrentSpec, WiretapFiltersStore,
    WiretapHttpTransactionStore, WiretapLinkCacheKey, WiretapLinkCacheStore,
    WiretapLocalStorage, WiretapReportChannel,
    WiretapSelectedTransactionStore,
    WiretapSpecStore, WiretapStaticChannel,
} from "@/model/constants";

declare global {
    interface Window {
        wiretapPort: any;
    }
}

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    private readonly _storeManager: BagManager;
    private readonly _httpTransactionStore: Bag<HttpTransaction>;
    private readonly _selectedTransactionStore: Bag<HttpTransaction>;
    private readonly _filtersStore: Bag<WiretapFilters>;
    private readonly _controlsStore: Bag<WiretapControls>;
    private readonly _linkCacheStore: Bag<Map<string, Map<string, HttpTransactionBase[]>>>;
    private readonly _specStore: Bag<string>;
    private readonly _bus: Bus;
    private readonly _wiretapChannel: Channel;
    private readonly _wiretapSpecChannel: Channel;
    private readonly _wiretapControlsChannel: Channel;
    private readonly _wiretapReportChannel: Channel;
    private readonly _wiretapConfigChannel: Channel;
    private readonly _staticNotificationChannel: Channel;
    private readonly _wiretapPort: string;
    private _transactionChannelSubscription: Subscription;
    private _specChannelSubscription: Subscription;
    private _configChannelSubscription: Subscription;
    private _staticChannelSubscription: Subscription;

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

        // extract port from session storage.
        this._wiretapPort = sessionStorage.getItem("wiretapPort");

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

        // filters store & subscribe to filter changes.
        this._filtersStore = this._storeManager.createBag(WiretapFiltersStore);

        // link cache store
        this._linkCacheStore =
            this._storeManager.createBag<Map<string, Map<string, HttpTransactionBase[]>>>(WiretapLinkCacheStore);

        // set up wiretap channels
        this._wiretapChannel = this._bus.createChannel(WiretapChannel);
        this._wiretapSpecChannel = this._bus.createChannel(SpecChannel);
        this._wiretapControlsChannel = this._bus.createChannel(WiretapControlsChannel);
        this._wiretapReportChannel = this._bus.createChannel(WiretapReportChannel);
        this._wiretapConfigChannel = this._bus.createChannel(WiretapConfigurationChannel);
        this._staticNotificationChannel = this._bus.createChannel(WiretapStaticChannel);

        // map local bus channels to broker destinations.
        this._bus.mapChannelToBrokerDestination(TopicPrefix + WiretapChannel, WiretapChannel);
        this._bus.mapChannelToBrokerDestination(QueuePrefix + SpecChannel, SpecChannel);
        this._bus.mapChannelToBrokerDestination(QueuePrefix + WiretapControlsChannel, WiretapControlsChannel);
        this._bus.mapChannelToBrokerDestination(QueuePrefix + WiretapReportChannel, WiretapReportChannel);
        this._bus.mapChannelToBrokerDestination(QueuePrefix + WiretapConfigurationChannel, WiretapConfigurationChannel);
        this._bus.mapChannelToBrokerDestination(TopicPrefix + WiretapStaticChannel, WiretapStaticChannel);

        // handle incoming messages on different channels.
        this._transactionChannelSubscription = this._wiretapChannel.subscribe(this.wireTransactionHandler());
        this._specChannelSubscription = this._wiretapSpecChannel.subscribe(this.specHandler());
        this._configChannelSubscription = this._wiretapConfigChannel.subscribe(this.configHandler());
        this._staticChannelSubscription = this._staticNotificationChannel.subscribe(this.staticHandler());


        // load previous transactions from local storage.
        this.loadHistoryFromLocalStorage().then((previousTransactions: Map<string, HttpTransaction>) => {
            // populate store with previous transactions.
            this._httpTransactionStore.populate(previousTransactions)

            // calculate counts from stored state.
            this.calculateMetricsFromState(previousTransactions);
        });

        // configure wiretap broker.
        const config = {
            brokerURL: 'ws://localhost:' + this._wiretapPort + '/ranch',
            heartbeatIncoming: 0,
            heartbeatOutgoing: 0,
            onConnect: () => {
                this.requestSpec();
            }
        }
        this._bus.connectToBroker(config);
    }

    calculateMetricsFromState(previousTransactions: Map<string, HttpTransaction>) {
        let requests = 0;
        let responses = 0;
        let violations = 0;
        let violated = 0.0
        if (previousTransactions) {
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
        }
        this.requestCount = requests;
        this.responseCount = responses;
        this.violationsCount = violations;
        this.violatedTransactions = violated;
        this.calcComplianceLevel();
    }

    requestSpec() {
        this._bus.publish({
            destination: "/pub/queue/specs",
            body: JSON.stringify({requestCommand: GetCurrentSpecCommand}),
        })
    }

    async loadHistoryFromLocalStorage(): Promise<Map<string, HttpTransaction>> {
        return localforage.getItem<Map<string, HttpTransaction>>(WiretapLocalStorage);
    }

    specHandler(): BusCallback<CommandResponse> {
        return (msg: CommandResponse) => {
            const decoded = atob(msg.payload.payload);
            this._specStore.set(WiretapCurrentSpec, decoded)
            localforage.setItem(WiretapCurrentSpec, decoded);
            this.requestUpdate();
        }
    }


    configHandler(): BusCallback<CommandResponse> {
        return (msg: CommandResponse) => {
           // todo: do something in here.
        }
    }


    staticHandler(): BusCallback<CommandResponse> {
        return (msg: CommandResponse) => {
            // todo: do something in here.
        }
    }

    wireTransactionHandler(): BusCallback {
        return (msg: CommandResponse) => {
            const wiretapMessage = msg.payload as HttpTransaction

            const constructedTransaction: HttpTransaction = new HttpTransaction();
            constructedTransaction.httpRequest = Object.assign(new HttpRequest(), wiretapMessage.httpRequest);
            constructedTransaction.id = wiretapMessage.id;
            constructedTransaction.requestValidation = wiretapMessage.requestValidation;
            constructedTransaction.responseValidation = wiretapMessage.responseValidation;

            // get global delay
            const controls = this._controlsStore.get(WiretapControlsKey)
            if (controls.globalDelay > 0) {
                constructedTransaction.delay = controls.globalDelay;
            }

            // get chain link cache
            const linkCache = this._linkCacheStore.get(WiretapLinkCacheKey);
            if (linkCache) {
                linkCache.forEach((value: Map<string, HttpTransactionBase[]>, key: string) => {
                    // check if a link has been detected.
                    if (constructedTransaction.httpRequest?.query?.includes(key)) {
                        constructedTransaction.containsChainLink = true;
                    }
                });
            }

            if (wiretapMessage.requestValidation && wiretapMessage.requestValidation.length > 0) {
                this.violatedTransactions += 0.5;
                this.violationsCount += wiretapMessage.requestValidation.length
            }

            if (wiretapMessage.httpResponse) {
                this.responseCount++;
                constructedTransaction.httpResponse = Object.assign(new HttpResponse(), wiretapMessage.httpResponse);
                if (wiretapMessage.responseValidation && wiretapMessage.responseValidation.length > 0) {
                    this.violatedTransactions += 0.5;
                    this.violationsCount += wiretapMessage.responseValidation.length
                }
            }

            const existingTransaction: HttpTransaction = this._httpTransactionStore.get(constructedTransaction.id)

            if (existingTransaction) {
                if (constructedTransaction.httpResponse) {
                    existingTransaction.httpResponse = constructedTransaction.httpResponse
                    existingTransaction.responseValidation = constructedTransaction.responseValidation
                    this._httpTransactionStore.set(existingTransaction.id, existingTransaction)
                }
            } else {
                this.requestCount++;
                constructedTransaction.timestamp = new Date().getTime();
                this._httpTransactionStore.set(constructedTransaction.id, constructedTransaction)
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
                this._specStore,
                this._filtersStore);
            this._transactionContainer = transaction;
        }
        let noSpec = false
        if (this._specStore) {
            const spec = this._specStore.get(WiretapCurrentSpec);
            if (!spec || spec == NoSpec) {
                noSpec = true
            }
        }

        if (noSpec) {
            return html`
            <wiretap-header
                    @wipeData=${this.wipeData}
                    requests="${this.requestCount}"
                    responses="${this.responseCount}"
                    violations="${this.violationsCount}"
                    violationsDelta="${this.violatedTransactions}"
                    compliance="${this.complianceLevel}"
                    noSpec>
            </wiretap-header>
            ${transaction}`
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