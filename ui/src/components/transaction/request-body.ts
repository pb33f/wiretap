import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import {HttpRequest} from "@/model/http_transaction";
import {
    ContentTypeFormEncoded,
    ContentTypeHtml,
    ContentTypeJSON, ContentTypeMultipartForm,
    ContentTypeOctetStream,
    ContentTypeXML, ExtractContentTypeFromRequest,
    FormPart
} from "@/model/extract_content_type";
import {unsafeHTML} from "lit/directives/unsafe-html.js";
import prismCss from "@/components/prism.css";
import Prism from "prismjs";
import 'prismjs/components/prism-json';
import 'prismjs/components/prism-xml-doc';
import 'prismjs/themes/prism-okaidia.css';
import sharedCss from "@/components/shared.css";
import {KVViewComponent} from "@pb33f/cowboy-components/components/kv-view/kv-view";
import {HttpPropertyViewComponent} from "@pb33f/cowboy-components/components/http-property-view/http-property-view";
import requestViewCss from "./request-body.css";
import {languages} from "monaco-editor";
import json = languages.json;

@customElement('request-body-view')
export class RequestBodyViewComponent extends LitElement {

    static styles = [prismCss, sharedCss, requestViewCss];

    private readonly _httpRequest: HttpRequest;

    constructor(req: HttpRequest) {
        super();
        this._httpRequest = req;
    }

    parseFormEncodedData(data: string): Map<string, string> {
        const map = new Map<string, string>();
        if (data) {
            const pairs = data.split('&');
            for (const pair of pairs) {
                const [key, value] = pair.split('=');
                map.set(decodeURI(key), decodeURI(value));
            }
        }
        return map;
    }

    render() {

        const req = this._httpRequest;
        const exct = ExtractContentTypeFromRequest(req)
        const ct = html` <span class="contentType">
            Content Type: <strong>${exct}</strong>
        </span>`;

        let jsonBody = '[unable to parse JSON body]';
        let parsedBody: any = req.requestBody;


        switch (exct) {
            case ContentTypeJSON:

                try {
                    parsedBody = JSON.parse(req.requestBody);
                    jsonBody = JSON.stringify(parsedBody, null, 2)
                    parsedBody = jsonBody;
                } catch (e) {
                    jsonBody += ": " + e.message;
                }


                return html`${ct}
                <pre><code>${unsafeHTML(Prism.highlight(jsonBody,
                        Prism.languages.json, 'json'))}</code></pre>`;

            case ContentTypeXML:
                return html`${ct}
                <pre><code>${unsafeHTML(Prism.highlight(req.requestBody,
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
                <pre><code>${unsafeHTML(Prism.highlight(req.requestBody,
                        Prism.languages.xml, 'xml'))}</code></pre>`;

            case ContentTypeFormEncoded:
                const kv = new KVViewComponent();
                kv.keyLabel = "Form Key";
                kv.data = this.parseFormEncodedData(req.requestBody);
                return html`${ct}${kv}`

            case ContentTypeMultipartForm:
                const formProps = new HttpPropertyViewComponent()
                formProps.propertyLabel = "Form Key";
                formProps.typeLabel = "Type";

                try {
                    parsedBody = JSON.parse(req.requestBody);
                    jsonBody = JSON.stringify(parsedBody, null, 2)
                } catch (e) {
                    jsonBody += ": " + e.message;
                }

                // extract pre-rendered form data from wiretap
                const parts: FormPart[] = parsedBody as FormPart[];
                for (const part of parts) {
                    if (part.value?.length > 0) {
                        part.type = 'field';
                    }
                    if (part.files?.length > 0) {
                        part.type = 'file';
                    }
                }

                formProps.data = parts;

                return html`${ct}${formProps}`

            default:
                return html`${ct}
                <pre>${req.requestBody}</pre>`
        }
    }

}

