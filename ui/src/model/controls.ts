import {HttpTransaction} from "@/model/http_transaction";
import {RanchUtils} from "@pb33f/ranch";

export class WiretapControls {
    globalDelay: number;
}

export class WiretapFilters {

    constructor() {
        this.filterMethod = {
            id: RanchUtils.genShortId(5),
            keyword: "",
        }
        this.filterKeywords = [];
        this.filterChain = [];
    }
    filterMethod: Filter;
    filterKeywords: Filter[];
    filterChain: Filter[];

}

export function AreFiltersActive(filters: WiretapFilters): boolean {
    if (filters.filterMethod.keyword.length > 0) {
        return true;
    }
    if (filters.filterKeywords.length > 0) {
        return true;
    }
    return filters.filterChain.length > 0;

}


export interface Filter {
    id?: string;
    keyword: string;
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