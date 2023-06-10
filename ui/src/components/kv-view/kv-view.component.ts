import {customElement, state, property} from "lit/decorators.js";
import {map} from "lit/directives/map.js";

import {LitElement, TemplateResult} from "lit";
import {html} from "lit";
import kvViewComponentCss from "./kv-view.component.css";

import Prism from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/themes/prism-okaidia.css';
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import prismCss from "@/components/prism.css";
import sharedCss from "@/components/shared.css";

@customElement('http-kv-view')
export class KVViewComponent extends LitElement {

    static styles = [prismCss, sharedCss, kvViewComponentCss];

    @property()
    keyLabel: string = 'Header';

    @property()
    valueLabel: string = 'Value';

    @state()
    private _data: Map<string, any>;

    constructor() {
        super();
        this._data = new Map<string, any>();
    }

    set data(value: Map<string, any>) {
        this._data = value;
    }

    render() {
        let headerData: TemplateResult
        if (this._data) {
            headerData = html`
                ${map(this._data, (i) => {
                    if (typeof i[1] === 'object') {
                        return html`
                            <tr>
                                <td><code>${i[0]}</code></td>
                                <td>
                                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(i[1]), Prism.languages['json'], 'json'))}</pre></code>
                                </td>
                            </tr>`
                    }

                    return html`
                        <tr>
                            <td><code>${i[0]}</code></td>
                            <td>${i[1]}</td>
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
            <div class="kv-table">
                <table>
                    <thead>
                    <tr>
                        <th>${this.keyLabel}</th>
                        <th>${this.valueLabel}</th>
                    </tr>
                    </thead>
                    <tbody>
                    ${headerData}
                    </tbody>
                </table>
            </div>
        `;

        const output = this._data?.size > 0 ? table : noData;

        return html`${output}`;
    }
}