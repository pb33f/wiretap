import {customElement, state, query} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {GlobalDelayChangedEvent, RequestReportEvent, WipeDataEvent} from "@/model/events";
import {property} from "lit/decorators.js";
import sharedCss from "@/components/shared.css";
import settingsComponentCss from "@/components/controls/settings.css";
import {SlInput} from "@shoelace-style/shoelace";

@customElement('wiretap-controls-settings')
export class WiretapControlsSettingsComponent extends LitElement {

    static styles = [sharedCss, settingsComponentCss]

    @property({type: Number})
    globalDelay: number;

    @query('#global-delay')
    globalDelayInput: SlInput;

    handleGlobalDelayChange() {
        this.dispatchEvent(new CustomEvent(GlobalDelayChangedEvent, {detail: this.globalDelayInput.value}))
    }

    wipeData() {
        this.dispatchEvent(new CustomEvent(WipeDataEvent))
    }

    sendReportRequest() {
        this.dispatchEvent(new CustomEvent(RequestReportEvent))
    }

    render() {
        return html`
            <label>Global API Delay (MS)</label>
            <sl-input id="global-delay" @sl-change=${this.handleGlobalDelayChange} value=${this.globalDelay}
                      placeholder="size" size="medium" type="number" id="global-delay">
                <sl-icon name="hourglass-split" slot="prefix"></sl-icon>
            </sl-input>
            <hr/>
            <sl-button @click=${this.wipeData} variant="danger" outline>Reset State</sl-button>
            <hr/>
            <sl-button @click=${this.sendReportRequest} outline>
                <sl-icon name="save" slot="prefix"></sl-icon>
                Download Session Data
            </sl-button>`

    }
}
