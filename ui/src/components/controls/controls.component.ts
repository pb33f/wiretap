import {customElement, query, state} from "lit/decorators.js";
import {LitElement} from "lit";
import {html} from "lit";
import {ControlsResponse, WiretapConfig, WiretapControls} from "@/model/controls";
import localforage from "localforage";
import {Bus, BusCallback, Channel, CommandResponse, GetBus, Message, Subscription} from "@pb33f/ranch";
import controlsComponentCss from "./controls.component.css";
import {SlDrawer, SlInput} from "@shoelace-style/shoelace";
import {RanchUtils} from "@pb33f/ranch";
import {GetBagManager, BagManager, Bag} from "@pb33f/saddlebag";
import {WipeDataEvent} from "@/model/events";
import sharedCss from "@/components/shared.css";
import {
    ChangeDelayCommand,
    WiretapControlsChannel, WiretapControlsKey,
    WiretapControlsStore,
    WiretapHttpTransactionStore
} from "@/model/constants";

@customElement('wiretap-controls')
export class WiretapControlsComponent extends LitElement {

    static styles = [sharedCss, controlsComponentCss]

    @state()
    private _controls: WiretapControls;

    private readonly _bus: Bus;

    @query('sl-drawer')
    drawer: SlDrawer;

    @query("#global-delay")
    delayInput: SlInput;

    @state()
    private _drawerOpen: boolean = false;

    private readonly _wiretapControlsSubscription: Subscription;
    private readonly _wiretapControlsChannel: Channel;
    private readonly _storeManager: BagManager;
    private readonly _controlsStore: Bag<WiretapControls>;

    constructor() {
        super();

        // get bus.
        this._bus = GetBus();
        this._storeManager = GetBagManager();
        this._controlsStore = this._storeManager.getBag(WiretapControlsStore);
        this._wiretapControlsChannel = this._bus.getChannel(WiretapControlsChannel);
        this._wiretapControlsSubscription = this._wiretapControlsChannel.subscribe(this.controlUpdateHandler());

        this.loadControlStateFromStorage().then((controls: WiretapControls) => {
            if (!controls) {
                this._controls = {
                    globalDelay: -1,
                }
            } else {
                this._controls = controls;
            }
            // get the delay from the backend.
            this.changeGlobalDelay(-1) // -1 won't update anything, but will return the current delay
        });
    }

    async loadControlStateFromStorage(): Promise<WiretapControls> {
        return localforage.getItem<WiretapControls>(WiretapControlsStore);
    }

    controlUpdateHandler(): BusCallback<CommandResponse> {
        return (msg: Message<CommandResponse<ControlsResponse>> ) => {
            const delay = msg.payload.payload.config.globalAPIDelay;
            const existingDelay = this._controls.globalDelay;

            if (delay == undefined) {
                // this means a reset back to 0.
                this._controls.globalDelay = 0;
            }

            if (delay != undefined && delay !== existingDelay) {
                this._controls.globalDelay = delay;
            }

            // update the store
            this._controlsStore.set(WiretapControlsKey, this._controls)
            localforage.setItem<WiretapControls>(WiretapControlsStore, this._controls);
        }
    }

    changeGlobalDelay(delay: number) {
        this._bus.getClient().publish({
            destination: "/pub/queue/controls",
            body: JSON.stringify(
                {
                    id: RanchUtils.genUUID(),
                    requestCommand: ChangeDelayCommand,
                    payload: {
                        delay: delay
                    }
                }
            ),
        });
    }

    openControls() {
        this.drawer.show();
    }

    closeControls() {
        this.drawer.hide()
    }

    handleGlobalDelayChange(event) {
        const delay = event.target.value;
        this.changeGlobalDelay(parseInt(delay));
    }

    wipeData() {
        this.dispatchEvent(new CustomEvent(WipeDataEvent, {
            bubbles: true,
            detail: WiretapHttpTransactionStore,
            composed: true,
        }));
    }

    render() {

        return html`
            <sl-button @click=${this.openControls} variant="default" size="medium" circle outline>
                <sl-icon name="gear" label="controls" class="gear"></sl-icon>
            </sl-button>
            <sl-drawer label="wiretap controls" class="drawer-focus">
                <label>Global API Delay (MS)</label>
                <sl-input @sl-change=${this.handleGlobalDelayChange} value=${this._controls?.globalDelay} placeholder="size" size="medium" type="number" id="global-delay">
                    <sl-icon name="hourglass-split" slot="prefix"></sl-icon>
                </sl-input>
                <hr />
                <sl-button @click=${this.wipeData} variant="danger" outline>Erase all HTTP Traffic</sl-button>
                <sl-button @click=${this.closeControls} slot="footer" variant="primary" outline>Close</sl-button>
            </sl-drawer>
        `
    }
}