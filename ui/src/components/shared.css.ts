import {css} from "lit";

export default css`

  hr {
    height: 1px;
    border-bottom: none;
    border-left: none;
    border-right: none;
    border-top: 1px dashed var(--secondary-color);
    margin-bottom: 20px;
    margin-top: 10px;
  }

  .empty-data {
    text-align: center;
    padding-top: 20px;
    color: var(--darker-font-color)
  }

  .empty-data .mute-icon {
    font-size: 100px;
    margin-bottom: 20px;
    color: var(--dark-font-color);
  }

  .empty-data .binary-icon {
    font-size: 100px;
    margin-bottom: 20px;
    color: var(--secondary-color);
  }

  .empty-data .up-icon {
    font-size: 100px;
    margin-bottom: 20px;
    color: var(--primary-color);
  }

  .empty-data .ok-icon {
    font-size: 100px;
    margin-bottom: 20px;
    color: var(--primary-color);
  }
  
  .empty-data.ok {
    color: var(--primary-color);
  }

  .empty-data.engage {
    padding-top: 90px;
    color: var(--primary-color);
  }

  .binary-data .binary-icon {
    font-size: 100px;
    margin-bottom: 20px;
    color: var(--primary-color);
  }

  sl-tag.method {
    width: 80px;
    text-align: center;
  }

  .method::part(base) {
    background: var(--background-color);
    border-radius: 0;
    text-align: center;
    font-family: var(--font-stack);
    width:100%;
  }

  .method::part(content) {
    border-radius: 0;
    text-align: center;
    width: 100%;
    display: inline-block;
  }
  
`