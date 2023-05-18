import {css} from "lit";

export default css`
  
  .transaction {
    border-bottom: 1px dashed var(--secondary-color-lowalpha);
    border-top: 1px dashed var(--background-color);
    padding: 5px;
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
  
  


`