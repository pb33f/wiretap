export const HttpTransactionSelectedEvent = "httpTransactionSelected";
export const ViolationLocationSelectionEvent = "violationLocationSelected";
export interface ViolationLocation {
    line: number;
    column: number;
}