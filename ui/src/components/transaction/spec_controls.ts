import {customElement, property, state} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {ToggleSpecificationEvent} from "@/model/events";
import spec_controlsCss from "@/components/transaction/spec_controls.css";

@customElement('spec-controls')
export class SpecControlsComponent extends LitElement {

    static styles = spec_controlsCss

    @state()
    specVisible: boolean;

    constructor() {
        super();
    }

    toggleSpec() {
        this.specVisible = !this.specVisible;
        this.dispatchEvent(new CustomEvent(ToggleSpecificationEvent, {
            detail: this.specVisible
        }));
    }

    render() {

        return html`
            <sl-button variant="default" size="small" @click="${this.toggleSpec}">
                <sl-icon slot="prefix" name="${this.specVisible? 'eye-slash' : 'eye'}"></sl-icon>
                ${this.specVisible ? 'OpenAPI' : 'OpenAPI'}
            </sl-button>`

    }
}
