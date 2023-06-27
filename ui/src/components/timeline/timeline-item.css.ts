import {css} from "lit";

export default css`
  :host {
    display: flex;
    position: relative;
  }

  .icon {
    width: 80px;
    min-height: 60px;
    position: relative;
  }

  .icon > .timeline {
    width: 40px;
    height: 100%;
    border-right: 2px solid var(--secondary-color);
    position: absolute;
  }

  .timeline-icon {
    width: 20px;
    height: 20px;
    border-radius: 15px;
    margin: 0 10px;
    text-align: center;
    font-size: 20px;
    position: absolute;
    top: 15px;
    background: var(--secondary-color-very-lowalpha);
  }

  .content {
    flex-grow: 2;
    height: 50px;
  }

  .request-time {
    font-size: 0.7em;
    color: var(--dark-font-color);
  }
  
`
