import {customElement, property} from "lit/decorators.js";
import {LitElement} from "lit";
import {html} from "lit";
import metricsCss from "./metrics.css";

@customElement('wiretap-header-metrics')
export class HeaderMetricsComponent extends LitElement {
    static styles = metricsCss;

    @property({type: Number})
    requests: number;

    @property({type: Number})
    responses: number;

    @property({type: Number})
    violations: number;

    @property({type: Number})
    violationsDelta: number;

    @property({type: Number})
    compliance: number;

    render() {
        return html`
            <wiretap-metric title="Requests" value="${this.requests}"></wiretap-metric>
            <wiretap-metric title="Responses" value="${this.responses}"></wiretap-metric>
            <wiretap-metric title="Violations" value="${this.violations}"></wiretap-metric>
            <wiretap-metric title="Compliance" value="${this.compliance}" postfix="%" end colorizeValue></wiretap-metric>
        `
    }
}