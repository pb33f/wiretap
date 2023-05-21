import {css} from "lit";

export default css`
  .error-icon {
    color: var(--warn-color);
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
  
  
  .how-to-fix {
    
  }
`