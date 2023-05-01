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


import {Client} from "@stomp/stompjs";


const client = new Client({
    brokerURL: 'ws://localhost:9090/ws',
    heartbeatIncoming: 0,
    heartbeatOutgoing: 0,
    onConnect: () => {

        client.subscribe('/topic/wiretap-broadcast', message =>
            console.log(`broadcast: ${message.body}`)
        );
    },
});

client.activate();

