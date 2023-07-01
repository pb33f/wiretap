import {css} from "lit";

export default css`
  :host {
    display: flex;
    position: relative;
  }

  .icon {
    width: 100px;
    min-height: 80px;
    position: relative;
  }

  .icon:first-child {
    min-height: 20px;
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
    top: 10px;
  }

  .content {
    flex-grow: 2;
    height: 50px;
  }

  .request-time {
    font-size: 0.7em;
    color: var(--dark-font-color);
    position: absolute;
    left: 50px;
    top: 15px;
    display: block;
  }
  
`
