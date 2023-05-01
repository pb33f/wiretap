import '../css/variables.css'
import '../css/pb33f.css'
import '../css/navigation.css'
import '../css/header.css'
import '../css/syntax.css'

import '@shoelace-style/shoelace/dist/themes/light.css';
import '@shoelace-style/shoelace/dist/themes/dark.css';

import './test';
import './wiretap';
import './components/header/header.component';
import {Bus, Channel, CreateBus, Subscription} from "./ranch/bus";
import {HttpTransaction} from "./model/http_transaction";
import {CreateStore, Store, StoreAllChangeSubscriptionFunction} from "./ranch/store";

const WiretapChannel =  "wiretap-broadcast";


const bus: Bus = CreateBus()

const config = {
    brokerURL: 'ws://localhost:9090/ws',
    heartbeatIncoming: 0,
    heartbeatOutgoing: 0,
}


const transactionStore: Store<HttpTransaction>  = CreateStore<HttpTransaction>()



const channel: Channel = bus.createChannel(WiretapChannel)

const sub: Subscription = channel.subscribe((msg)  => {
    const httpTransaction: HttpTransaction = msg.payload as HttpTransaction
    const existingTransaction: HttpTransaction = transactionStore.get(httpTransaction.id)
    if (existingTransaction) {
        if (httpTransaction.httpResponse) {
            existingTransaction.httpResponse = httpTransaction.httpResponse
            transactionStore.set(existingTransaction.id, existingTransaction)
        }
    } else {
        transactionStore.set(httpTransaction.id, httpTransaction)
    }
})
transactionStore.onAllChanges( (key: string, value: HttpTransaction) => {
   if (value.httpResponse) {
       console.log(JSON.parse(value.httpResponse.responseBody))
   }
})

bus.mapChannelToBrokerDestination("/topic/" + WiretapChannel, WiretapChannel)

bus.connectToBroker(config)






