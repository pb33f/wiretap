import {css} from "lit";

export default css`
  
  
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







`