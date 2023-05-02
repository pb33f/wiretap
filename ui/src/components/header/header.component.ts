import {customElement} from "lit/decorators.js";
import {html, LitElement} from "lit";
import headerCss from "./header.css";

@customElement('wiretap-header')
export class HeaderComponent extends LitElement {
    static styles = headerCss;
    render() {
        return html`<header class="site-header">
            <div class="logo">
                <span class="caret">$</span>
                <span class="name"><a href="https://pb33f.io?ref=wiretap-ui">wiretap</a></span>
            </div>
            <div class="header-space">
            </div>
        </header>`
    }
}