import {customElement, state, property} from "lit/decorators.js";
import {map} from "lit/directives/map.js";

import {LitElement, TemplateResult} from "lit";
import {html} from "lit";

import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/themes/prism-okaidia.css';
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import prismCss from "@/components/prism.css";
import sharedCss from "@/components/shared.css";
import propertyViewComponentCss from "./property-view.css";
import {FormDataEntry, FormPart} from "@/model/extract_content_type";

export interface Property extends FormPart {
}

@customElement('http-property-view')
export class PropertyViewComponent extends LitElement {

    static styles = [prismCss, sharedCss, propertyViewComponentCss];

    @property()
    propertyLabel: string = 'Property';

    @property()
    typeLabel: string = 'Type';

    @property()
    valueLabel: string = 'Value';

    @state()
    private _data: Property[];

    constructor() {
        super();
        this._data = [];
    }

    set data(value: Property[]) {
        this._data = value;
    }

    render() {
        let headerData: TemplateResult
        if (this._data) {
            headerData = html`
                ${map(this._data, (i) => {
                    if (i.type !== 'field') {
                        return html`
                            <tr>
                                <td><code>${i.name}</code></td>
                                <td><code>${i.type}</code></td>
                                <td>${i.files.map((f) => {
                                    return html`
                                        <span class="file-name">${f.name}</span>
                                        <hr/>
                                        ${Object.keys(f.headers).map((key, index) => {
                                            return html`
                                                <span class="file-header">${key}
                                                    : <strong>${f.headers[key]}</strong></span>`
                                        })}</td>
                                        </tr>`
                                })}
                                </td>`
                    }

                    return html`
                        <tr>
                            <td><code>${i.name}</code></td>
                            <td>${i.type}</td>
                            <td>${i.value}</td>
                        </tr>`
                })}
            `
        }

        const noData: TemplateResult = html`
            <div class="empty-data">
                <sl-icon name="mic-mute" class="mute-icon"></sl-icon>
                <br/>
                No data extracted
            </div>`;

        const table: TemplateResult = html`
            <div class="prop-type-table">
                <table>
                    <thead>
                    <tr>
                        <th>${this.propertyLabel}</th>
                        <th>${this.typeLabel}</th>
                        <th>${this.valueLabel}</th>
                    </tr>
                    </thead>
                    <tbody>
                    ${headerData}
                    </tbody>
                </table>
            </div>
        `;

        const output = this._data?.length > 0 ? table : noData;

        return html`${output}`;
    }
}