import {customElement, property, state} from "lit/decorators.js";
import {html, LitElement, TemplateResult} from "lit";
import {HttpTransaction} from "@/model/http_transaction";
import transactionComponentCss from "@/components/transaction/transaction-item.css";
import Prism from 'prismjs'
import 'prismjs/components/prism-javascript'
import 'prismjs/themes/prism-okaidia.css'
import {HttpTransactionSelectedEvent} from "@/model/events";
import sharedCss from "@/components/shared.css";
import {TransactionLinkCache} from "@/model/link_cache";


@customElement('http-transaction-item')
export class HttpTransactionItemComponent extends LitElement {

    static styles = [sharedCss, transactionComponentCss]

    public _linkCache: TransactionLinkCache;

    @state()
    _httpTransaction: HttpTransaction

    @state()
    _active = false;

    private _processing = false;

    @property({type: Boolean})
    hideControls = false;

    constructor(httpTransaction: HttpTransaction, linkCache: TransactionLinkCache) {
        super();
        this._linkCache = linkCache;
        this._httpTransaction = httpTransaction;
    }

    get transactionId(): string {
        return this._httpTransaction.id
    }

    get active(): boolean {
        return this._active;
    }

    set httpTransaction(value: HttpTransaction) {
        this._httpTransaction = value;
    }

    get httpTransaction(): HttpTransaction {
        return this._httpTransaction;
    }

    disable() {
        this._active = false;
    }

    setActive(): void {
        if (this._active) {
            return
        }
        this._active = true;
        const options = {
            detail: this._httpTransaction,
            bubbles: true,
            composed: true,
        };
        this.dispatchEvent(
            new CustomEvent<HttpTransaction>(HttpTransactionSelectedEvent, options)
        );
    }


    render() {
        Prism.highlightAll();
        const req = this._httpTransaction?.httpRequest;
        const resp = this._httpTransaction?.httpResponse;
        this._processing = req && !resp;

        let tClass = "transaction";
        if (this._active) {
            tClass += " active";
        }

        let statusIcon = html``

        if (this._processing) {
            statusIcon = html`<sl-spinner class="spinner"></sl-spinner>`
        } else {
            if (req && resp) {
                if (this._httpTransaction.requestValidation?.length > 0 ||
                    this._httpTransaction.responseValidation?.length > 0) {
                    statusIcon = html`<sl-icon name="exclamation-circle" class="invalid"></sl-icon>`
                } else {

                    if (this._httpTransaction.httpResponse?.statusCode >= 400) {
                        statusIcon = html`<sl-icon name="x-square" class="failed"></sl-icon>`
                    } else {
                        statusIcon = html`<sl-icon name="check-lg" class="valid"></sl-icon>`
                    }

                }
            }
        }

        let delay: TemplateResult;
        if (this._httpTransaction.delay > 0) {
            delay = html`<div class="delay">${this._httpTransaction.delay}ms <sl-icon name="hourglass-split" ></sl-icon></div>`
        }

        let chainLink: TemplateResult;

        if (this._httpTransaction.containsChainLink) {
            const matches = this._linkCache.findLinks(this.httpTransaction);
            let total = matches.length;
            let totalMatches = 0;
            matches.forEach((m) => {
                totalMatches += m.siblings.length;
            });

            chainLink = html`
                <sl-tooltip>
                    <div slot="content">
                        <strong>${total}</strong> parameter(s), <strong>${totalMatches}</strong>
                        matching request(s)
                    </div>
                    <div class="chain"><sl-icon name="link-45deg"></sl-icon></div>
                </sl-tooltip>
               `
        }

        let respTime = 0;
        let reqTime = 0;
        if (resp) {
            respTime = resp.timestamp;
        }
        if (req) {
            reqTime = req.timestamp;
        }

        let totalTime = respTime - reqTime;
        if (totalTime < 0) {
            totalTime = 0;
        }

        return html`
            <div class="${tClass}" @click="${this.setActive}">
                <header>
                   <pb33f-http-method method="${req.method}"></pb33f-http-method>
                    ${decodeURI(req.path)}
              
                </header>
                ${delay}
                <div class="request-time">
                    ${(totalTime > 10000) ? html`${(totalTime/1000).toFixed(1)}s` : 
                            html`${totalTime>0 ? totalTime : null}${totalTime>0 ? 'ms' : null}`} 
                    <sl-icon name="arrow-left-right"></sl-icon>
                </div>
                <div class="transaction-status">
                    ${this.hideControls? '' : chainLink}
                    ${statusIcon}
                </div>
            </div>`
    }
}