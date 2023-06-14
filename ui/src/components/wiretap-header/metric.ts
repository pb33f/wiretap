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

    @property({type: Boolean})
    colorizeValue: boolean;

    render() {
        return html`
            <div class="metric ${this.end ? 'end' : null}">
                <span class="title">${this.title}</span><br/>
                <span class="value ${this.colorizeValue ? this.calcStyle() : null}">${this.value}${this.postfix}</span>
            </div>
        `;
    }

    calcStyle(): string {
        if (this.value <= 20)
            return "error";
        if (this.value > 20 && this.value <= 30)
            return "big-warning";
        if (this.value > 30 && this.value <= 50)
            return "warning";
        if (this.value > 50 && this.value <= 70)
            return "light-warning";
        if (this.value > 70 && this.value <= 95)
            return "ok";
        if (this.value > 95)
            return "good";
    }
}