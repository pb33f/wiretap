
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



import './css/variables.css'
import './css/pb33f.css'
import './css/header.css'
import './css/syntax.css'

import './components/wiretap-header/header.component';
import './components/transaction/transaction-container.component';
import './components/transaction/transaction-view.component';
import './components/violation/violation.component';
import './components/kv-view/kv-view.component';
import './components/editor/editor.component';
import './components/wiretap-header/metrics.component';
import './components/wiretap-header/metric.component';
import './components/controls/controls.component';
import './components/controls/settings.component';
import './components/controls/filters.component';


import './model/http_transaction';
import './wiretap';


// configure shoelace
import {setBasePath} from '@shoelace-style/shoelace/dist/utilities/base-path.js';

// Set the base path to the folder you copied Shoelace's assets to
//setBasePath('/assets/shoelace');
setBasePath('/shoelace');
