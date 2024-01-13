

import {css} from "lit";

export default css`

    sl-icon.report {
        font-size: 1.4rem;
        display: inline-block;
    }

    label {
        display: block;
        padding-bottom: 10px;
    }

    hr {
        margin-top: 30px;
        margin-bottom: 30px;
    }

    sl-select::part(combobox) {
        border-radius: 0;
    }

    sl-input::part(base) {
        border-radius: 0;
    }

    sl-button::part(base) {
        font-family: var(--font-stack), monospace;
        border: 1px dashed;
        border-radius: 0;
    }
    
    

`