import {css} from "lit";

export default css`

    :host {
        position: absolute;
        right: 16px;
        top: 10px;
        height: 30px;
        width: 90px;
    }

    sl-button::part(base) {
        font-family: var(--font-stack), monospace;
        border: 1px dashed;
        border-radius: 0;

    }

    sl-icon {
        font-size: 1.4em
    }




`