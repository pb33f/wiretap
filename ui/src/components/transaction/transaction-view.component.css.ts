import {css} from "lit";

export default css`
  
  
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
  }
  
  h3 {
    margin-top: 5px;
    margin-bottom: 15px;
    font-family: var(--font-stack-heading);
    color: var(--primary-color);
  }

  .violation-badge::part(base) {
    font-size: 0.5rem;
    margin-left: 5px;
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
    color: var(--warn-300);
  }
  .http500 {
    color: var(--error-color);
  }
  
  .response-code {
    font-size: 1.2em;
  }
  
`