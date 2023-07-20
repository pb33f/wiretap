import {customElement, state, property} from "lit/decorators.js";
import {map} from "lit/directives/map.js";

import {LitElement, TemplateResult} from "lit";
import {html} from "lit";

import sharedCss from "@/components/shared.css";
import propertyViewComponentCss from "@/components/property-view/property-view.css";
import {SchemaValidationFailure} from "@/model/http_transaction";
import {ViolationLocation, ViolationLocationSelectionEvent} from "@/model/events";
import violationCss from "@/components/violation/violation.css";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import prismCss from "@/components/prism.css";
import Prism from "prismjs";
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-yaml';
import 'prismjs/themes/prism-okaidia.css';
import 'prismjs/plugins/line-numbers/prism-line-numbers.css';
import {SlRadioButton, SlRadioGroup, SlSwitch} from "@shoelace-style/shoelace";

enum SchemaType {
    SCHEMA = 'schema',
    OBJECT = 'object'
}

@customElement('violation-details-view')
export class ViolationDetailsComponent extends LitElement {

    static styles = [sharedCss, violationCss, prismCss, propertyViewComponentCss];

    @state()
    private selectedViolationView: SchemaType = SchemaType.SCHEMA;

    @state()
    private showSchemaObjects = false;


    @state()
    private _data: SchemaValidationFailure[];

    constructor(data: SchemaValidationFailure[]) {
        super();
        this._data = data;
    }

    set data(value: SchemaValidationFailure[]) {
        this._data = value;
    }

    jumpToLocation(line: number, column: number) {
        return () => {
            this.dispatchEvent(new CustomEvent<ViolationLocation>(ViolationLocationSelectionEvent, {
                bubbles: true,
                composed: true,
                detail: {
                    line: line,
                    column: column,
                }
            }))
        }
    }

    protected firstUpdated() {
        const rb: NodeListOf<SlRadioButton> = this.renderRoot.querySelectorAll('sl-radio-button') as NodeListOf<SlRadioButton>;

        rb.forEach(
            (rb: SlRadioButton) => {
                rb.disabled = !this.showSchemaObjects;
            }
        )

        // console.log(sw?.checked);

        // const sw: SlSwitch = this.renderRoot.querySelector('sl-switch') as SlSwitch;
        // this.showSchemaObjects = sw.checked;

    }

    schemaViolationObjectSwitch() {
        const radio = this.renderRoot.querySelector('.schema-radio-group') as SlRadioGroup;
        this.selectedViolationView = radio.value as SchemaType;
    }

    showObjectsSwitch() {
        const sw: SlSwitch = this.renderRoot.querySelector('sl-switch') as SlSwitch;
        this.showSchemaObjects = sw.checked;

        const rb: NodeListOf<SlRadioButton> = this.renderRoot.querySelectorAll('sl-radio-button') as NodeListOf<SlRadioButton>;

        rb.forEach(
            (rb: SlRadioButton) => {
                rb.disabled = !this.showSchemaObjects;
            }
        );
    }

    render() {
        let violationDetails: TemplateResult
        let radioGroup: TemplateResult;
        let schemaDataView: TemplateResult;

        radioGroup = html`
            <sl-radio-group class="schema-radio-group" name="a" value=${SchemaType.SCHEMA}
                            size="small" @sl-change="${this.schemaViolationObjectSwitch}">
                <sl-radio-button class="schema-radio-button" size="small" value='${SchemaType.SCHEMA}'
                                 ${!this.showSchemaObjects ? html`disabled` : html`disabled`}>Validation Schema
                </sl-radio-button>
                <sl-radio-button class="schema-radio-button" size="small" value='${SchemaType.OBJECT}'
                                 ${!this.showSchemaObjects ? html`disabled` : html`disabled`}>Validated Object
                </sl-radio-button>
            </sl-radio-group>`


        if (this._data) {

            violationDetails = html`
                ${map(this._data, (i) => {
                    
                    const formatted = Prism.highlight(i.referenceSchema,
                            Prism.languages.json, 'json');
                    
                    let schemaView: TemplateResult;
                    if (this.selectedViolationView === SchemaType.SCHEMA) {
        
                        const formattedCode = formatted
                                .split('\n')
                                .map((line, num) => `<span class="${((num+1) == i.line) ?
                                        'line-active' : ''}"><span class="line-num ${((num+1) == i.line) ? 
                                        'line-active' : ''}">${(num + 1).toString().padStart(4, ' ')}.</span> ${line}</span>`)
                                .join('\n');
                        
                        schemaView = html`
                            <div class="schema-violation-object">
                                <pre class="line-numbers"><code>${unsafeHTML(formattedCode)}</code></pre>
                            </div>`
                    }
                    if (this.selectedViolationView === SchemaType.OBJECT) {
                        schemaView = html`
                            <div class="schema-violation-object">
                                <pre><code>${unsafeHTML(Prism.highlight(i.referenceObject,
                                        Prism.languages.json, 'json'))}</code></pre>
                            </div>`
                    }

                    if (this.showSchemaObjects) {
                        schemaDataView = html`
                            <tr>
                                <td style="width: 75%" colspan="3">
                                    <section class="schema-violation-objects">
                                        ${schemaView}
                                    </section>
                                </td>
                            </tr>
                        `
                    }

                    return html`
                        <tr>
                            <td>${i.line}</td>
                            <td>${i.location}</td>
                            <td style="width: 75%">${i.reason}</td>
                        </tr>
                        ${schemaDataView}
                    `
                })}
            `
        }

        const noData: TemplateResult = html`
            <div class="empty-data">
                <sl-icon name="mic-mute" class="mute-icon"></sl-icon>
                <br/>
                No schema violations found.
            </div>`;

        const table: TemplateResult = html`
            <section class="schema-data-switch">
                <div class="schema-data-switch-input">
                    <sl-switch size="small" @sl-change="${this.showObjectsSwitch}">Show validation schema and
                        object
                    </sl-switch>
                </div>
                <div class="schema-type-select">
                    ${radioGroup}
                </div>
            </section>
            <div class="prop-type-table">
                <table>
                    <thead>
                    <tr>
                        <th>Line</th>
                        <th>XPath</th>
                        <th>Reason</th>
                    </tr>
                    </thead>
                    <tbody>
                    ${violationDetails}
                    </tbody>
                </table>
            </div>
        `;

        const output = this._data?.length > 0 ? table : noData;
        return html`${output}`;
    }

    updated() {
        Prism.highlightAll();
    }

}