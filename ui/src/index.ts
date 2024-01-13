import '@shoelace-style/shoelace/dist/themes/light.css';
import '@shoelace-style/shoelace/dist/themes/dark.css';
import '@shoelace-style/shoelace/dist/components/tag/tag.js';
import '@shoelace-style/shoelace/dist/components/badge/badge.js';
import '@shoelace-style/shoelace/dist/components/tab-panel/tab-panel.js';
import '@shoelace-style/shoelace/dist/components/tab-group/tab-group.js';
import '@shoelace-style/shoelace/dist/components/tab/tab.js';
import '@shoelace-style/shoelace/dist/components/split-panel/split-panel.js';
import '@shoelace-style/shoelace/dist/components/details/details.js';
import '@shoelace-style/shoelace/dist/components/icon/icon.js';
import '@shoelace-style/shoelace/dist/components/spinner/spinner.js';
import '@shoelace-style/shoelace/dist/components/drawer/drawer.js';
import '@shoelace-style/shoelace/dist/components/button/button.js';
import '@shoelace-style/shoelace/dist/components/input/input.js';
import '@shoelace-style/shoelace/dist/components/dropdown/dropdown.js';
import '@shoelace-style/shoelace/dist/components/menu/menu.js';
import '@shoelace-style/shoelace/dist/components/menu-item/menu-item.js';
import '@shoelace-style/shoelace/dist/components/divider/divider.js';
import '@shoelace-style/shoelace/dist/components/select/select.js';
import '@shoelace-style/shoelace/dist/components/option/option.js';
import '@shoelace-style/shoelace/dist/components/tooltip/tooltip.js';
import '@shoelace-style/shoelace/dist/components/switch/switch.js';
import '@shoelace-style/shoelace/dist/components/radio-button/radio-button.js';
import '@shoelace-style/shoelace/dist/components/radio-group/radio-group.js';
import '@shoelace-style/shoelace/dist/components/icon-button/icon-button.js';


import '@pb33f/cowboy-components/cowboy-components.css';
import '@pb33f/cowboy-components/components/header/header.js';
import '@pb33f/cowboy-components/components/theme-switcher/theme-switcher.js';
import '@pb33f/cowboy-components/components/timeline/timeline.js';
import '@pb33f/cowboy-components/components/timeline/timeline-item.js';
import '@pb33f/cowboy-components/components/http-method/http-method.js';
import '@pb33f/cowboy-components/components/attention-box/attention-box.js';
import '@pb33f/cowboy-components/components/mailing-list/mailing-list.js';
import '@pb33f/cowboy-components/components/render-operation-path/render-operation-path.js';


// css
// import './css/variables.css'
// import './css/pb33f.css'
// import './css/header.css'
// import './css/syntax.css'

// wiretap components
import './components/wiretap-header/header';
import './components/transaction/transaction-container';
import './components/transaction/transaction-view';
import './components/transaction/response-body';
import './components/transaction/spec_controls';
import './components/violation/violation';
import './components/editor/editor';
import './components/wiretap-header/metrics';
import './components/wiretap-header/metric';
import './components/controls/controls';
import './components/controls/settings.component';
import './components/controls/filters.component';

// models
import './model/http_transaction';

// boot.
import './wiretap';

// Set the base path to the folder you copied Shoelace's assets to
import {setBasePath} from '@shoelace-style/shoelace/dist/utilities/base-path.js';
//setBasePath('/shoelace');
setBasePath('/assets/shoelace');
