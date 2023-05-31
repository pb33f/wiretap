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
  
  .method::part(base) {
    background: var(--background-color);
    border-radius: 0;
    text-align: center;
    width:100%;
  }

  .method::part(content) {
    border-radius: 0;
    text-align: center;
    width: 100%;
    display: inline-block;
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
  
  sl-tag {
    width: 80px;
    text-align: center;
  }
  
  .delay {
    width: 70px;
    margin: 5px 10px 0 10px;
    color: var(--secondary-color);
  }
  .delay sl-icon {
    vertical-align: bottom;
    font-size: 21px;
    color: var(--dark-font-color);
  }
  
`