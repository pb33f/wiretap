import {css} from "lit";

export default css`
  a, a:visited, a:active {
    text-decoration: none;
    color: var(--primary-color);
  }

  a:hover {
    color: var(--primary-color);
    text-decoration: underline;
  }
  
  wiretap-controls {
    width: 100px;
    height: 55px;
    position: absolute;
    right: 2px
  }
  
  
  header.site-header {
    display: flex;
    height: 57px;
    flex-direction: row;
    width: 100vw;
    z-index: 1;
    background-color: var(--background-color);
  }

  header.site-header > .logo {
    width: 170px;
    padding: 10px 0 10px 10px;
    border-bottom: 1px dashed var(--secondary-color);
    height: 35px;
  }

  header.site-header > .logo .caret {
    font-size: 1.6em;
    font-weight: lighter;
    color: var(--secondary-color)
  }

  header.site-header > .logo .name {
    font-size: 1.7em;
    font-weight: bold;
    color: var(--primary-color);
    text-shadow:  0 0 10px var(--primary-color-lowalpha), 0 0 10px rgba(251, 169, 255, 0.06);
  }

  header.site-header > .logo .name > a {
    text-decoration: none;
  }

  header.site-header > .logo .name > a:hover {
    text-decoration: underline;

  }

  header.site-header > .logo .name > a:active {
    text-decoration: underline;
    color: var(--secondary-color);
    text-shadow:  0 0 5px var(--secondary-color-text-shadow), 0 0 10px rgba(251, 169, 255, 0.06);
  }

  header.site-header > .logo .name::after {
    content: "";
    -webkit-animation: cursor .8s infinite;
    animation: cursor .8s infinite;
    background: var(--primary-color);
    border-radius: 0;
    display: inline-block;
    height: 0.9em;
    margin-left: 0.2em;
    width: 3px;
    bottom: -2px;
    position: relative;
  }
  
  header .header-space {
    height: 55px;
    flex-grow: 2;
    border-bottom: 1px dashed var(--secondary-color);
  }

  @-webkit-keyframes cursor {
    0% {
      opacity: 0;
    }

    50% {
      opacity: 1;
    }

    to {
      opacity: 0;
    }
  }

  @keyframes cursor {
    0% {
      opacity: 0;
    }

    50% {
      opacity: 1;
    }

    to {
      opacity: 0;
    }
  }


  @media only screen and (max-width: 1000px) {

    .generated-timestamp {
      display: none;
    }
    .main-content {
      border-bottom: none;
    }

  }   
    `