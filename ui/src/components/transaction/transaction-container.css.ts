import {css} from "lit";

export default css`
  
  :host {
    position: relative;
  }
  
  .bottom-panel {
    position: relative;
    height: 100%;
    width: 100%;
  }
  
  .transactions {
    padding: 20px;
  }

  .split-panel-divider sl-split-panel {
    --divider-width: 2px;
    height: calc(100vh - 57px); 
    --min: 150px; 
    --max: calc(100% - 400px);
  }

  .split-panel-divider sl-split-panel.editor-split {
    height: 100%;  
    --min: 300px; 
    --max: calc(100% - 250px);
  }
  
  .split-panel-divider sl-split-panel::part(divider) {
    background-color: var(--secondary-color);
  }

  .split-panel-divider sl-icon {
    position: absolute;
    border-radius: var(--sl-border-radius-small);
    background: var(--background-color);
    color: var(--secondary-color);
    border: 1px dashed var(--secondary-color);
    //height: 20px;
    //width: 40px;
    padding: 0.125rem 0.5rem;
    //padding: 0;
  }

  .split-panel-divider sl-icon.grip-vertical {
    position: absolute;
    border-radius: var(--sl-border-radius-small);
    background: var(--background-color);
    color: var(--secondary-color);
    border: 1px dashed var(--secondary-color);
    //height: 20px;
    //width: 40px;
    padding: 0.5rem 0.125rem;
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
    height: calc(100% - 2px);
    width: 100%;
    overflow: auto;
    position: relative;
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


