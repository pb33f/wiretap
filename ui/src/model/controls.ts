
export class WiretapControls {
    globalDelay: number;
}

export interface ControlsResponse {
    config: WiretapConfig;
}

export interface WiretapConfig {
    redirectHost:   string;
    port:           string;
    monitorPort:    string;
    globalAPIDelay: number;
}