import {customElement, property} from "lit/decorators.js";
import {html, LitElement} from "lit";
import headerCss from "./header.css";


@customElement('wiretap-header')
export class HeaderComponent extends LitElement {
    static styles = headerCss;

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
            <header class="site-header">
                <div class="logo">
                    <span class="caret">$</span>
                    <span class="name"><a href="https://pb33f.io?ref=wiretap-ui">wiretap</a></span>
                </div>
                <div class="header-space">
                    <wiretap-header-metrics
                            requests="${this.requests}"
                            responses="${this.responses}"
                            violations="${this.violations}"
                            violationsDelta="${this.violationsDelta}"
                            compliance="${this.compliance}">
                    </wiretap-header-metrics>
                   
                </div>
                <wiretap-controls></wiretap-controls>
            </header>`
    }
}