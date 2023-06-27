import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import wiretapTimelineItemCss from "./timeline-item.css";

@customElement('wiretap-timeline-item')
export class TimelineItemComponent extends LitElement {

    static styles = wiretapTimelineItemCss

    constructor() {
        super();
    }

    render() {

        return html`
            <div class="icon">
                <div class="timeline"></div>
                <div class="timeline-icon">
                    <slot name="icon"></slot>
                </div>
            </div>
            <div class="content">
                <slot name="time" class="request-time"></slot>
                <slot name="content" class="timeline-content"></slot>
            </div>`
    }
}