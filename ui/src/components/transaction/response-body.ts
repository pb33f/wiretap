import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpResponse} from "@/model/http_transaction";
import {
    ContentTypeHtml,
    ContentTypeJSON,
    ContentTypeOctetStream,
    ContentTypeXML,
    ExtractContentTypeFromResponse
} from "@/model/extract_content_type";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import prismCss from "@/components/prism.css";
import Prism from "prismjs";
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-xml-doc';
import 'prismjs/themes/prism-okaidia.css';
import sharedCss from "@/components/shared.css";
import responseBodyCss from "./response-body.css";

@customElement('response-body-view')
export class ResponseBodyViewComponent extends LitElement {

    static styles = [prismCss, sharedCss, responseBodyCss];
    private readonly _httpResponse: HttpResponse;

    constructor(resp: HttpResponse) {
        super();
        this._httpResponse = resp
    }

    render() {
        const exct = ExtractContentTypeFromResponse(this._httpResponse)
        const ct = html` <span class="contentType">
            Content Type: <strong>${exct}</strong>
        </span>`;

        switch (exct) {
            case ContentTypeXML:
                return html`
                    <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(this._httpResponse.responseBody), null, 2),
                            Prism.languages.xml, 'xml'))}</code></pre>`;
            case ContentTypeOctetStream:
                return html`${ct}
                <div class="empty-data">
                    <sl-icon name="file-binary" class="binary-icon"></sl-icon>
                    <br/>
                    [ binary data will not be rendered ]
                </div>`;
            case ContentTypeHtml:
                return html`${ct}
                <pre><code>${unsafeHTML(Prism.highlight(this._httpResponse.responseBody,
                        Prism.languages.xml, 'xml'))}</code></pre>`;

            default:
                return html`${ct}
                <pre><code>${unsafeHTML(Prism.highlight(JSON.stringify(JSON.parse(this._httpResponse.responseBody), null, 2),
                        Prism.languages.json, 'json'))}</code></pre>`;
        }
    }

}

