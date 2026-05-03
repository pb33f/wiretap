
import {css} from "lit";

export default css`

    :host {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        height: 55px;
    }

    .controls-toolbar {
        display: flex;
        align-items: center;
        justify-content: flex-end;
        gap: 8px;
        height: 100%;
        white-space: nowrap;
    }

    pb33f-theme-switcher {
        position: static;
        margin-top: 0;
        flex: 0 0 auto;
    }
    
    .filter-control {
        position: relative;
        display: inline-flex;
        align-items: center;
        flex: 0 0 auto;
    }
    
    sl-icon-button {
        margin: 0;
        font-size: 1.3rem;
        padding-top: 0;
        flex: 0 0 auto;
    }

    sl-drawer::part(panel) {
        background-color: var(--background-color);
        border-left: 1px dashed var(--secondary-color);
    }

    sl-drawer::part(body) {
        background-color: var(--background-color);
    }

    sl-drawer::part(header) {
        background-color: var(--background-color);
    }

    sl-drawer::part(footer) {
        background-color: var(--background-color);
    }

    sl-button::part(base) {
        font-family: var(--font-stack), monospace;
        margin-top: 6px;
        border: 1px dashed;
        border-radius: 0;
    }

    .filters-badge::part(base) {
        font-size: 0.5rem;
        background-color: var(--primary-color);
        position: absolute;
        top: 7px;
        right: -2px;
        z-index: 1;
    }

`
