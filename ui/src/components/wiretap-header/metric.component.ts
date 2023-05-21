import {customElement, property} from "lit/decorators.js";
import {LitElement} from "lit";
import {html} from "lit";
import metricCss from "./metric.css";

@customElement('wiretap-metric')
export class HeaderMetricComponent extends LitElement {
    static styles = metricCss;

    @property({type: Boolean})
    end: boolean;

    @property()
    title: string;

    @property()
    value: number;

    @property()
    postfix: string;

    render() {
        return html`
            <div class="metric ${this.end ? 'end' : null}">
                <span class="title">${this.title}</span><br/>
                <span class="value">${this.value}${this.postfix}</span>
            </div>
        `;
    }
}