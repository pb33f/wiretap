import './css/variables.css'
import './css/pb33f.css'
import './css/header.css'
import './css/syntax.css'

import '@shoelace-style/shoelace/dist/themes/light.css';
import '@shoelace-style/shoelace/dist/themes/dark.css';
import '@shoelace-style/shoelace/dist/components/tag/tag.js';
import '@shoelace-style/shoelace/dist/components/tab-panel/tab-panel.js';
import '@shoelace-style/shoelace/dist/components/tab-group/tab-group.js';
import '@shoelace-style/shoelace/dist/components/tab/tab.js';
import '@shoelace-style/shoelace/dist/components/split-panel/split-panel.js';
import '@shoelace-style/shoelace/dist/components/details/details.js';

import './components/wiretap-header/header.component';
import './components/transaction/transaction-container.component';
import './components/transaction/transaction-view.component';
import '@/components/kv-view/kv-view.component';
import './ranch/bus';
import './model/http_transaction';
import './ranch/store';
import './wiretap';


// configure shoelace
import {setBasePath} from '@shoelace-style/shoelace/dist/utilities/base-path.js';

// Set the base path to the folder you copied Shoelace's assets to
setBasePath('/shoelace');



// transactionStore.onAllChanges( (key: string, value: HttpTransaction) => {
//    if (value.httpResponse) {
//        console.log(JSON.parse(value.httpResponse.responseBody))
//    }
// })








