import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import timelineCss from "@/components/timeline/timeline.css";

@customElement('wiretap-timeline')
export class TimelineComponent extends LitElement {

    static styles = timelineCss;

    constructor() {
        super();
    }

    render() {
        return html`
            <div class="start">
                <div class="ball-start"></div>
            </div>
            <slot></slot>
            <div class="end">
                <div class="ball-end"></div>
            </div>
        `
    }
}