import {css} from "lit";

export default css`
  .transactions {
    padding: 20px;
  }

  .split-panel-divider sl-split-panel {
    --divider-width: 1px;
  }

  .split-panel-divider sl-split-panel::part(divider) {
    background-color: var(--secondary-color);
  }

  .split-panel-divider sl-icon {
    position: absolute;
    border-radius: var(--sl-border-radius-small);
    background: var(--divider-handle-background);
    color: var(--primary-color);
    border: 1px dashed var(--secondary-color);
    //height: 20px;
    //width: 40px;
    padding: 0.125rem 0.5rem;
    //padding: 0;
  }

  .split-panel-divider sl-split-panel::part(divider):focus-visible {
    background-color: var(--secondary-color);
  }

  .split-panel-divider sl-split-panel:focus-within sl-icon {
    background-color: var(--sl-color-primary-600);
    color: var(--sl-color-neutral-0);
  }
  
  .transactions-container {
    height: 100%; 
    align-items: start; 
    overflow-y: auto
  }

  .transaction-view-container {
    height: 100%;
    align-items: start;
    overflow: auto;
  }
  
  

  .transactions-container::-webkit-scrollbar {
    width: 8px;
  }

  .transactions-container::-webkit-scrollbar-track {
    background-color: var(--invert-font-color);
  }

  .transactions-container::-webkit-scrollbar-thumb {
    box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    background: var(--secondary-color-lowalpha);
  }
  
  
  
`


