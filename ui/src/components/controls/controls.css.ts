
import {css} from "lit";

export default css`

    pb33f-theme-switcher {
        position: absolute;
        margin-top: 8px;
        right: 2px;
    }
    
    
    sl-icon-button {
        margin: 3px auto;
        font-size: 1.3rem;
        padding-top: 6px;
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
        top: 5px;
        left: 25px;
    }

`