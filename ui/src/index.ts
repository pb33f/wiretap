import './css/variables.css'
import './css/pb33f.css'
import './css/header.css'
import './css/syntax.css'

import '@shoelace-style/shoelace/dist/themes/light.css';
import '@shoelace-style/shoelace/dist/themes/dark.css';

import './wiretap';
import './components/header/header.component';
import './components/transaction/transaction-container.component';
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








