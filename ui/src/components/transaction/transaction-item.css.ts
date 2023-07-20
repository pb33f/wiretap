import {css} from "lit";

export default css`
  
  .transaction {
    border-bottom: 1px dashed var(--secondary-color-lowalpha);
    border-top: 1px dashed var(--background-color);
    padding: 5px;
    display: flex;
    justify-content: space-between;
  }
  
  header {
    flex-grow: 5;
  }
  
  .transaction:hover{
    background: var(--transaction-background-color-hover);
    cursor: pointer;
    border-bottom: 1px dashed var(--secondary-color);
    border-top: 1px dashed var(--secondary-color);
  }

  .transaction:active {
    background: var(--transaction-background-color-active);
    border-bottom: 1px dashed var(--primary-color);
    border-top: 1px dashed var(--primary-color);
    color: var(--primary-color);
  }

  
  .transaction.active {
    background: var(--kv-table-header-background);
    border-bottom: 1px dashed var(--primary-color);
    border-top: 1px dashed var(--primary-color);
    color: var(--primary-color);
  }
  


  .tab::part(base) {
    font: var(--font-stack);
  }

  .tab.secondary::part(base) {
    font: var(--font-stack);
  }
  
  .spinner {
    --indicator-color: var(--secondary-color); 
    font-size: 25px; 
    --track-width: 3px;
    margin-top: 2px;
    margin-right: 12px;
  }
  
  .invalid {
    color: var(--error-color);
    font-size: 25px;
    margin-top: 2px;
    margin-right: 12px;
  }

  .valid {
    color: var(--terminal-green);
    font-size: 25px;
    margin-top: 2px;
    margin-right: 12px;
  }

  .failed {
    color: var(--dark-font-color);
    font-size: 25px;
    margin-top: 2px;
    margin-right: 12px;
  }



  .delay {
    text-align: right;
    min-width: 60px;
    max-width: 70px;
    font-size: 12px;
    margin: 8px 10px 0 10px;
    color: var(--secondary-color);
  }
  .delay sl-icon {
    vertical-align: bottom;
    font-size: 15px;
    color: var(--dark-font-color);
  }

  .chain {
    width: 20px;
    margin: 5px 10px 0 10px;
    color: var(--secondary-color);
  }
  .chain sl-icon {
    vertical-align: bottom;
    font-size: 21px;
    color: var(--dark-font-color);
  }

  .chain sl-icon:hover {
    color: var(--terminal-yellow);
  }
  
  .transaction-status {
    display: flex; 
  }
  
  .request-time {
    text-align: right;
    min-width: 60px;
    max-width: 70px;
    font-size: 12px;
    margin: 8px 10px 0 10px;
    color: var(--secondary-color);
  }

  .request-time sl-icon {
    vertical-align: bottom;
    font-size: 15px;
    color: var(--dark-font-color);
  }
  

  
`