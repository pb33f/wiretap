import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";

@customElement('wiretap-application')
export class WiretapComponent extends LitElement {

    render() {
        return html`<wiretap-header></wiretap-header>`
    }

}