import {css} from "lit";

export default css`
  
  .transaction {
    border-bottom: 1px dashed var(--secondary-color-lowalpha);
    border-top: 1px dashed var(--background-color);
    padding: 5px;
    display: flex;
    justify-content: space-between;
  }
  
  .transaction:hover{
    background: var(--kv-table-header-background);
    cursor: pointer;
    border-bottom: 1px dashed var(--secondary-color);
    border-top: 1px dashed var(--secondary-color);
  }

  .transaction:active {
    border-bottom: 1px dashed var(--primary-color);
    border-top: 1px dashed var(--primary-color);
    color: var(--primary-color);
  }

  
  .transaction.active {
    border-bottom: 1px dashed var(--primary-color);
    border-top: 1px dashed var(--primary-color);
    color: var(--primary-color);
  }
  
  .method::part(base) {
    background: var(--background-color);
    border-radius: 0;
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
    color: var(--warn-color);
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
  
`