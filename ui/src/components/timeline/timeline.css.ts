import {css} from "lit";

export default css`
  
  wiretap-timeline-item {
    width: 100%;
    margin-bottom: 20px;
  } 
  
  .start {
    position: relative;
    width: 100%;
    height: 20px;
  }
  
  .ball-start {
    width: 15px;
    height: 15px;
    border-radius: 10px;
    background-color: var(--secondary-color);
    position: absolute;
    left: 33px;
    top: 5px;
    z-index: 10;
  }

  .ball-end {
    width: 15px;
    height: 15px;
    border-radius: 10px;
    background-color: var(--secondary-color);
    position: absolute;
    left: 33px;
    top: 0;
    z-index: 10;
    margin-bottom: 20px;
  }
  
  .end {
    position: relative;
    width: 100%;
    height: 20px;
  }
  
`