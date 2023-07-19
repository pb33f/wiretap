import {customElement, state, property} from "lit/decorators.js";
import {html} from "lit";
import {LitElement, TemplateResult} from "lit";

import violationComponentCss from "./violation.css";
import {SchemaValidationFailure, ValidationError} from "@/model/http_transaction";
import {ViolationLocation, ViolationLocationSelectionEvent} from "@/model/events";

import sharedCss from "@/components/shared.css";
import {Property} from "@/components/property-view/property-view";
import {ViolationDetailsComponent} from "@/components/violation/violation-details";

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

    jumpToViolationLocation(violation: SchemaValidationFailure) {
        this.dispatchEvent(new CustomEvent<ViolationLocation>(ViolationLocationSelectionEvent, {
            bubbles: true,
            composed: true,
            detail: {
                line: violation.line,
                column: violation.column,
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

        let schemaViolations: TemplateResult = null;
        if (this.violation.validationErrors?.length > 0) {
            const violationsDetails = new ViolationDetailsComponent(this.violation.validationErrors)
            schemaViolations = html`
                <h3>Schema Violations:</h3>
                <section class="schema-violations">
                    ${violationsDetails}
                </section>`

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
               ${schemaViolations}
                <hr/>
                <p class="reason">${this.violation?.reason}</p>
                ${howToFix}
            </sl-details>
        `
    }


}