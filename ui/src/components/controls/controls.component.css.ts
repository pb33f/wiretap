
import {css} from "lit";

export default css`
  sl-button {
   margin:5px auto

  }
  sl-icon.gear {
    font-size: 1.7rem;
    padding-top:6px;
  }
  
  label {
    display: block;
    padding-bottom: 10px;
  }
  
  
  sl-drawer::part(panel) {
    background-color: var(--background-color);
    border-left: 1px dashed var(--secondary-color);
  }
  sl-drawer::part(body) {
    background-color: var(--background-color);
  }
  sl-drawer::part(header) {
    background-color: var(--background-color);
  }
  sl-drawer::part(footer) {
    background-color: var(--background-color);
  }
  
  sl-button::part(base) {
    font-family: var(--font-stack);
  }
`