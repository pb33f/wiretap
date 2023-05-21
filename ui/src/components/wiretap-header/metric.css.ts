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
  }
  
  .value {
    font-size: 1.3rem;
  }
  
`