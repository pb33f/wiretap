import {customElement, query, state} from "lit/decorators.js";
import {LitElement} from "lit";
import {html} from "lit";
import {WiretapControls} from "@/model/controls";
import localforage from "localforage";
import {ChangeDelayCommand, WiretapControlsStore} from "@/wiretap";
import {Bus, Channel, CreateBus} from "@/ranch/bus";
import controlsComponentCss from "./controls.component.css";
import {SlDrawer} from "@shoelace-style/shoelace";

@customElement('wiretap-controls')
export class WiretapControlsComponent extends LitElement {

    static styles = controlsComponentCss

    @state()
    private _controls: WiretapControls;

    private readonly _bus: Bus;

    @query('sl-drawer')
    drawer: SlDrawer;

    @state()
    private _drawerOpen: boolean = false;

    constructor() {
        super();

        // get bus.
        this._bus = CreateBus();

        this.loadControlStateFromStorage().then((controls: WiretapControls) => {
            if (!controls) {
                this._controls = {
                    globalDelay: 0,
                }
                localforage.setItem<WiretapControls>(WiretapControlsStore, this._controls);
            } else {
                this._controls = controls;
            }
        });
    }

    async loadControlStateFromStorage(): Promise<WiretapControls> {
        return localforage.getItem<WiretapControls>(WiretapControlsStore);
    }

    changeGlobalDelay(delay: number) {
        console.log('yeeeeeeee har', delay)
        this._bus.getClient().publish({
            destination: "/pub/queue/wiretap",
            body: JSON.stringify({requestCommand: ChangeDelayCommand, payload: delay}),
        })
    }

    openControls() {
        this.drawer.show();
    }

   closeControls() {
        this.drawer.hide()
    }

    handleGlobalDelayChange(event: CustomEvent) {
        const delay = event.detail.value;
        this.changeGlobalDelay(delay);
    }

    render() {
        console.log(this._controls);
        return html`
            <sl-button @click=${this.openControls} variant="default" size="medium" circle outline>
                <sl-icon name="gear" label="controls" class="gear"></sl-icon>
            </sl-button>
        <sl-drawer label="wiretap controls" class="drawer-focus">
            <label>Global API Delay (MS)</label>
            <sl-input @sl-change=${this.handleGlobalDelayChange} placeholder="size" size="medium" type="number">
                <sl-icon name="hourglass-split" slot="prefix"></sl-icon>
            </sl-input>
            <sl-button @click=${this.closeControls} slot="footer" variant="primary" outline>Close</sl-button>
        </sl-drawer>
        `
    }

}