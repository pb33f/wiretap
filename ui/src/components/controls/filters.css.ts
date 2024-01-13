
import {css} from "lit";

export default css`

    .label-on-left {
        --label-width: 5rem;
        --gap-width: 1rem;
    }

    .label-on-left + .label-on-left {
        margin-top: var(--sl-spacing-medium);
    }

    .label-on-left::part(form-control) {
        display: grid;
        grid: auto / var(--label-width) 1fr;
        gap: var(--sl-spacing-3x-small) var(--gap-width);
        align-items: center;
        font-family: var(--font-stack);
    }

    .label-on-left::part(form-control-label) {
        text-align: right;
    }

    .label-on-left::part(form-control-help-text) {
        grid-column-start: 2;
    }

    hr {
        margin-bottom: 20px;
        margin-top: 20px;
    }

    .keywords {
        border: 1px dashed var(--secondary-color-dimmer);
        padding: 5px;
        min-height: 40px;
        margin-top: 20px;
    }

    .chains {
        border: 1px dashed var(--primary-color-lowalpha);
        padding: 5px;
        min-height: 40px;
        margin-top: 20px;
    }

    .keyword {
        transition: var(--sl-transition-fast) opacity;
    }

    .keyword::part(base) {
        font-family: var(--font-stack), monospace;
        border: 1px dashed var(--secondary-color);
        margin-bottom: 5px;
        border-radius: 0;
    }

    .chain {
        transition: var(--sl-transition-fast) opacity;
    }

    .chain::part(base) {
        font-family: var(--font-stack), monospace;
        border: 1px dashed var(--primary-color);
        margin-bottom: 5px;
        border-radius: 0;
    }

    p {
        margin-top: 20px;
        margin-bottom: 20px;
    }

    .chain-input::part(base) {
        --sl-input-border-color: var(--primary-color);
    }
    
    sl-select::part(combobox) {
        border-radius: 0;
    }

    sl-input::part(base) {
        border-radius: 0;
    }
`