export const HttpTransactionSelectedEvent = "httpTransactionSelected";
export const ViolationLocationSelectionEvent = "violationLocationSelected";
export const GlobalDelayChangedEvent = "globalDelayChanged";
export const RequestReportEvent = "requestReport";

export const ToggleSpecificationEvent = "toggleSpecification";

export interface ViolationLocation {
    line: number;
    column: number;
}

export const WipeDataEvent = "wipeData";