import {css} from "lit";

export default css`
  .prop-type-table > table > thead > tr > th {
    font-family: var(--font-stack-paragraph);
    background-color: var(--table-header-background-solid);
    text-align: left;
    padding: 10px;
    color: var(--darker-font-color);
    font-weight: normal;
    border-bottom: 1px dashed var(--secondary-color-dimmer);
  }

  .prop-type-table > table > thead > tr > th:first-child {
    background: var(--kv-table-header-background-reversed);
    text-align: right;
    color: var(--secondary-color);
    padding-right: 10px;
  }

  .prop-type-table > table > thead > tr > th:nth-child(2) {
    text-align: center;
    color: var(--font-color);
  }

  .prop-type-table > table > thead > tr > th:last-child {
    background: var(--kv-table-header-background);
  }

  .prop-type-table > table > tbody > tr > td {
    font-family: var(--font-stack-paragraph);
    color: var(--darker-font-color);
    padding: 10px 0 10px 10px;
    border-bottom: 1px dashed var(--secondary-color-dimmer);
  }

  .prop-type-table > table > tbody > tr > td:first-child > code {
    color: var(--primary-color);
    font-weight: bold;
    font-size: 1em;
  }

  .prop-type-table > table > tbody > tr > td:first-child {
    width: 120px;
    text-align: right;
    padding-right: 10px;
    border-right: 1px dashed var(--secondary-color-dimmer);
  }
  
  .prop-type-table > table > tbody > tr > td:nth-child(2) {
    width: 10px;
    color: var(--font-color);
    font-family: var(--font-stack);
    font-size: 0.9em;
    font-style: italic;
    text-align: center;
    padding-right: 10px;
    border-right: 1px dashed var(--secondary-color-dimmer);
  }
  
  table {
    width: 100%;
    border-spacing: 0;
  }
  
  table > tbody > tr > td > a {
    text-decoration: underline;
  }

  table > tbody > tr > td > code {
    color: var(--secondary-color);
    font-weight: normal;
    font-size: 1em;
  }
  
  .file-item {
    display: block;
  }

  .file-name {
    font-weight: bold;
    display: block;
  }
  
  .file-header {
    color: var(--dark-font-color);
    display: inline-block;
  }
  
  .file-header > strong {
    color: var(--darker-font-color);
  
  }
  
  hr {
    margin-bottom: 10px;
  }
  
`