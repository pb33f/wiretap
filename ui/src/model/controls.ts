import {HttpTransaction} from "@/model/http_transaction";

export class WiretapControls {
    globalDelay: number;
}

export interface ControlsResponse {
    config: WiretapConfig;
}

export interface ControlsResponse {
    config: WiretapConfig;
}

export interface ReportResponse {
    transactions: HttpTransaction[];
}


export interface WiretapConfig {
    redirectHost:   string;
    port:           string;
    monitorPort:    string;
    globalAPIDelay: number;
}