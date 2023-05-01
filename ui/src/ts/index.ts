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

