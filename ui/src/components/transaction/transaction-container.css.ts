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

 
  
  .transactions-container {
    height: 100%; 
    align-items: start; 
    overflow-y: auto
  }

  .transaction-view-container {
    height: calc(100% - 2px);
    width: 100%;
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
  
  http-transaction-view {
    height: 100%;
  }
  
  
`


