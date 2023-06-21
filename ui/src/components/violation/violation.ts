import {customElement, state, property} from "lit/decorators.js";
import {html} from "lit";
import {LitElement, TemplateResult} from "lit";

import violationComponentCss from "@/components/violation/violation.css";
import {ValidationError} from "@/model/http_transaction";
import {ViolationLocation, ViolationLocationSelectionEvent} from "@/model/events";

import sharedCss from "@/components/shared.css";

@customElement('wiretap-violation-view')
export class ViolationViewComponent extends LitElement {

    static styles = [sharedCss, violationComponentCss];

    @property()
    violation: ValidationError;

    constructor() {
        super();
    }

    jumpToSpecLocation() {
        this.dispatchEvent(new CustomEvent<ViolationLocation>(ViolationLocationSelectionEvent, {
            bubbles: true,
            composed: true,
            detail: {
                line: this.violation.specLine,
                column: this.violation.specColumn
            }
        }))
    }

    render() {

        let howToFix, specMeta: TemplateResult
        if (this.violation?.howToFix) {
            howToFix = html`
                <h3>How to fix this violation:</h3>
                <p class="how-to-fix">${this.violation.howToFix}</p>`
        }

        if (this.violation?.specLine >= 0) {
            specMeta = html`
                <div class="location-meta">
                    Line: <span class="jump-spec" @click='${this.jumpToSpecLocation}'>${this.violation?.specLine}</span>
                    Column: <span class="jump-spec"
                                  @click='${this.jumpToSpecLocation}'>${this.violation?.specColumn}</span>
                </div>`
        }

        return html`
            <sl-details class="violation">
                <header slot="summary">
                    <sl-icon name="exclamation-square" class="error-icon"></sl-icon>
                    <strong>${this.violation?.message}</strong>
                </header>
                <div class="violation-meta">
                    <div>
                        Type:
                        <sl-tag size="small" class="validation-type">${this.violation?.validationType}</sl-tag>
                        /
                        <sl-tag size="small" class="validation-subtype">${this.violation?.validationSubType}</sl-tag>
                    </div>
                    ${specMeta}
                </div>
                <hr/>
                <p class="reason">${this.violation?.reason}</p>
                ${howToFix}
            </sl-details>
        `
    }


}