import {css} from "lit";

export default css`
  
  code,
  pre {
    color: var(--font-color);
    background: none;
    font-family: var(--font-stack);
    font-size: 1em;
    text-align: left;
    white-space: pre;
    word-spacing: normal;
    word-break: normal;
    word-wrap: normal;
    line-height: 1.5;
    tab-size: 4;
    hyphens: none;
  }

  /* Code blocks */

  pre[class*="language-"] {
    padding: 1em;
    margin: .5em 0;
    overflow: auto;
    border-radius: 0.3em;
  }


  /* Inline code */

  :not(pre) > code[class*="language-"] {
    padding: .1em;
    border-radius: .3em;
    white-space: normal;
  }

  .token.comment,
  .token.prolog,
  .token.doctype,
  .token.cdata {
    color: #7f7f7f;
  }

  .token.punctuation {
    color: #7f7f7f;
  }

  .token.namespace {
    opacity: .7;
  }

  .token.property,
  .token.tag,
  .token.constant,
  .token.symbol,
  .token.deleted {
    color: var(--primary-color);
  }

  .token.boolean,
  .token.number {
    color: var(--primary-color);
  }

  .token.selector,
  .token.attr-name,
  .token.string,
  .token.char,
  .token.builtin,
  .token.inserted {
    color: var(--secondary-color);
  }

  .token.operator,
  .token.entity,
  .token.url,
  .language-css .token.string,
  .style .token.string,
  .token.variable {
    color: var(--tertiary-color);
  }

  .token.atrule,
  .token.attr-value,
  .token.function,
  .token.class-name {
    color: var(--terminal-green);
  }

  .token.keyword {
    color: var(--primary-color);
  }

  .token.regex,
  .token.important {
    color: #fd971f;
  }

  .token.important,
  .token.bold {
    font-weight: bold;
  }

  .token.italic {
    font-style: italic;
  }

  .token.entity {
    cursor: help;
  }
  
  
  .method::part(base) {
    background: var(--background-color);
    border-radius: 0;
  }

  .tab::part(base) {
    font: var(--font-stack);
  }

  .tab-secondary::part(base) {
    font: var(--font-stack);
  }

  .tab-secondary::part(base) {
    --active: var(--secondary-color);
  }


  .secondary-tabs {
    --indicator-color: var(--secondary-color);
    font: var(--font-stack);
  }





  pre {
    border: none;
    border-left: 2px solid var(--secondary-color);
    border-top: 1px dashed var(--secondary-color-lowalpha);
    border-bottom: 1px dashed var(--secondary-color-lowalpha);
    padding: 10px 0 10px 10px;
    margin-top: 0;
    margin-bottom: 20px;
    font-size: 14px;

    background-color: rgba(0, 0, 0, 0);
    background-image: linear-gradient(to right, #171d25, var(--background-color));
    max-height: 500px;
  }
  
  pre code {
    display: block;
    overflow-y: auto;
    max-height: 500px;
    
  }


`