import {css} from "lit";

export default css`
  .error-icon {
    color: var(--error-color);
    font-size: 1.5em;
    vertical-align: bottom;
  }

  p.reason {
    margin-top: 0;
  }

  .violation::part(base) {
    background: none;
    border: dashed 1px var(--secondary-color-dimmer);
  }


  h3 {
    margin-top: 20px;
    margin-bottom: 10px;
    font-family: var(--font-stack-heading);
  }

  .violation-meta {
    display: flex;
    justify-content: space-between;
    color: var(--darker-font-color)
  }

  .location-meta {
    font-size: 0.9em;
    padding-top: 5px;
  }

  .jump-spec {
    color: var(--primary-color);
  }

  .jump-spec:hover {
    text-decoration: underline;
    cursor: pointer;
  }


  .validation-type::part(base) {
    background: none;
    border: dashed 1px var(--secondary-color-dimmer);
    color: var(--secondary-color)
  }

  .validation-subtype::part(base) {
    background: none;
    border: dashed 1px var(--secondary-color-dimmer);
    color: var(--secondary-color)
  }

  .schema-violation-objects {
    display: flex;
    margin-top: 5px;
    width: 100%
  }

  .schema-violation-objects .schema-violation-object {
    font-size: 0.7rem;
    flex-grow: 1;
    word-wrap: break-word;
    overflow-y: auto;
    font-style: normal;
    text-align: left;
    width: 100%
  }

  .schema-violation-object pre {
    white-space: pre-wrap;
  }

  .schema-violation-object pre > code {
    white-space: pre-wrap;
  }

  .line-num {
    color: var(--dark-font-color);
  }

  .line-active {
    color: red;
  }

  .schema-violation-objects .schema-violation-object h4 {
    font-size: 0.8rem;
    color: var(--primary-color);
    font-family: var(--mono-font-stack);
    font-weight: bold;
    text-transform: initial;
    margin-top: 0;
    margin-bottom: 0;
  }


  .schema-violation-objects .schema-violation-object::-webkit-scrollbar {
    width: 8px;
  }

  .schema-violation-objects .schema-violation-object::-webkit-scrollbar-track {
    background-color: var(--invert-font-color);
  }

  .schema-violation-objects .schema-violation-object::-webkit-scrollbar-thumb {
    box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    background: var(--secondary-color-lowalpha);
  }


  .schema-radio-button::part(base) {
    --sl-color-primary-600: var(--secondary-color);
  }

  .schema-radio-button::part(button) {
    font-family: var(--mono-font-stack);
  }

  .schema-violations {
  }

  .schema-violations::-webkit-scrollbar {
    width: 8px;
  }

  .schema-violations::-webkit-scrollbar-track {
    background-color: var(--invert-font-color);
  }

  .schema-violations::-webkit-scrollbar-thumb {
    box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    background: var(--secondary-color-lowalpha);
  }

  .schema-data-switch-input {
    width: 50%;
    padding-top: 3px;
  }

  .schema-type-select {
    text-align: right;
    width: 50%
  }

  .schema-radio-group {

  }

  a.schema-violation-link {
    color: var(--primary-color);
    font-family: var(--mono-font-stack);
    font-size: 0.8rem;
  }

  a.schema-violation-link:visited {
    color: var(--primary-color);
  }

  .how-to-fix {

  }

  .schema-data-switch {
    display: flex;
    text-align: left;
    margin-bottom: 10px;
    height: 30px;
    font-family: var(--mono-font-stack);
  }

`