import {css} from "lit";

export default css`
  
  :host {
    height: 100%;
  }
  
  .method::part(base) {
    background: var(--background-color);
    border-radius: 0;
  }

  .tab::part(base) {
    font: var(--font-stack);
  }

  #violation-tab::part(base) {
    width: 140px;
  }

  .tab-secondary::part(base) {
    font: var(--font-stack);
  }

  .secondary-tabs {
    --indicator-color: var(--secondary-color);
    font: var(--font-stack);
  }
  
  .tab-panel {
    padding: 0 10px 0 10px;
    height: 100%;
  }
  
  h3 {
    margin-top: 5px;
    margin-bottom: 15px;
    font-family: var(--font-stack-heading), monospace;
    color: var(--primary-color);
  }

  .violation-badge::part(base) {
    font-size: 0.5rem;
    margin-left: 5px;
    --sl-color-danger-600: var(--error-color);
  }
  
  h2 {
    margin:0;
    padding:0;
    font-size: 100px;
  }

  .http200 {
    color: var(--ok-color)
  }
  .http400 {
    color: var(--sl-color-warning-600);
  }
  .http500 {
    color: var(--error-color);
  }
  
  pre {
    max-width: calc(100vw - 135px);
    overflow-x: auto;
  }
  
  .chain-panel-divider {
    height: 100%
  }

  .chain-panel-divider sl-split-panel {
    --min: 150px; --max: calc(100% - 150px);
    
  }
  
  .chain-container {

    max-height: 300px;
    overflow-y: auto;
  }

  .chain-container::-webkit-scrollbar {
    width: 8px;
  }

  .chain-container::-webkit-scrollbar-track {
    background-color: var(--invert-font-color);
  }

  .chain-container::-webkit-scrollbar-thumb {
    box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    background: var(--secondary-color-lowalpha);
  }
  
  .empty-data.no-chain {
    margin-bottom: 20px;
  }
  
  .empty-data.select-chain {
    margin-bottom: 20px;
    color: var(--darker-font-color);
  }
  
  .chain-view-container {
    border-top: 1px solid var(--secondary-color);
  }
  
  .chain-panel-divider sl-split-panel::part(divider):focus-visible {
    background-color: var(--secondary-color);
  }

  .chain-panel-divider sl-split-panel:focus-within sl-icon {
    background-color: var(--sl-color-primary-600);
    color: var(--sl-color-neutral-0);
  }
  
  sl-split-panel {
    height: 100%;
  }
  
  .kv-overview {
    display: flex;
  }
  
  http-kv-view {
    flex-grow: 1;
  }
  
  .chain-time {
    text-align: left;
    width: 50%;
    padding-left: 20px;
    margin-top: 0;
    padding-top: 0;
    border-left: 1px dashed var(--secondary-color-lowalpha)
  }
  
  .chain-time-icon {
    font-size: 30px;
    vertical-align: baseline;
    padding-right: 10px;
    color: var(--secondary-color);
  }

  .request-chain-time-title {
    font-size: 0.7rem;
    color: var(--dark-font-color);
    padding-bottom: 10px;
  }
  
  .total-time {
    font-size: 2rem;
  }
  
  .time-value {
    color: var(--primary-color);
  }
  
  .dropped-header {
    color: var(--error-color);
  }
  
  .response-code {
    font-size: 0.8rem;
  }
  
`