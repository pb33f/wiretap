import {css} from "lit";

export default css`
  .metric {
    display: block;
    height: 55px;
    width: 130px;
    padding-left: 20px;
    padding-right: 20px;
    border-left: 1px dashed var(--secondary-color);
  }
  
  .end {
    border-right: 1px dashed var(--secondary-color);
  }
  
  .title {
    font-size: 0.7rem;
    color: var(--dark-font-color)
  }
  
  .value {
    font-size: 1.3rem;
  }
  
  .error {
    color: var(--error-color)
  }
  
  .big-warning {
    color: var(--warn-400);
  }
  
  .warning {
    color: var(--warn-200);
  }

  .light-warning {
    color: var(--ok-400);
  }

  .ok {
    color: var(--ok-300);
  }

  .ok {
    color: var(--terminal-green);
  }
`