import {css} from "lit";

export default css`
  table {
    width: 100%;
    border-spacing: 0;
  }

  .kv-table > table > thead > tr > th {
    font-family: var(--font-stack-paragraph);
    background: var(--kv-table-header-background);
    text-align: left;
    padding: 10px;
    color: var(--darker-font-color);
    font-weight: normal;
    border-bottom: 1px dashed var(--secondary-color-dimmer);
  }
  
  .kv-table > table > thead > tr > th:first-child {
    font-weight: bold;
    font-family: var(--font-stack-paragraph);
    background: var(--kv-table-header-background-reversed);
    text-align: right;
    color: var(--secondary-color);
    padding-right: 10px;
  }

  .kv-table > table > tbody > tr > td {
    font-family: var(--font-stack-paragraph);
    color: var(--darker-font-color);
    padding: 10px 0 10px 10px;
    border-bottom: 1px dashed var(--secondary-color-dimmer);
  }

  .kv-table > table > tbody > tr > td:first-child > code {
    color: var(--primary-color);
    font-weight: bold;
    font-size: 1em;
  }

  .kv-table > table > tbody > tr > td:first-child {
    width: 180px;
    text-align: right;
    padding-right: 10px;
    border-right: 1px dashed var(--secondary-color-dimmer);
  }

  table > tbody > tr > td > a {
    text-decoration: underline;
  }

  table > tbody > tr > td > code {
    color: var(--secondary-color);
    font-weight: normal;
    font-size: 1em;
  }
    

`