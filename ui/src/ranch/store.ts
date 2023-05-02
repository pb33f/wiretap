import {Subscription} from "./bus";
export type StoreValueSubscriptionFunction<T> = (value: T) => void;
export type StoreAllChangeSubscriptionFunction<T> = (key: string, value: T) => void;

export interface Store<T> {
    set(key: string, value: T): void;
    get(key: string): T;
    subscribe(key: string, callback: StoreValueSubscriptionFunction<T>): Subscription;
    onAllChanges(callback: StoreAllChangeSubscriptionFunction<T>): Subscription;
}

export function CreateStore<T>(): Store<T> {
    return new store<T>();
}

class store<T> {
    private _values: Map<string, T>;
    private _subscriptions: Map<string, StoreValueSubscriptionFunction<T>[]>;
    private _allChangesSubscriptions: StoreAllChangeSubscriptionFunction<T>[];

    constructor() {
        this._values = new Map<string, T>();
        this._subscriptions = new Map<string, StoreValueSubscriptionFunction<T>[]>()
        this._allChangesSubscriptions = [];
    }

    set(key: string, value: T): void {
        this._values.set(key, value);
        if (this._subscriptions.has(key)) {
            this._subscriptions.get(key).forEach((cb) => cb(value));
        }
        if(this._allChangesSubscriptions.length > 0) {
            this._allChangesSubscriptions.forEach((cb) => cb(key, value));
        }
    }

    get(key: string): T {
        return this._values.get(key);
    }
    
    subscribe(key: string, callback: StoreValueSubscriptionFunction<T>): Subscription {
        if (!this._subscriptions.has(key)) {
            this._subscriptions.set(key, [callback]);
        } else {
            const existingSubscriptions: StoreValueSubscriptionFunction<T>[] = this._subscriptions.get(key);
            this._subscriptions.set(key, [...existingSubscriptions, callback]);
        }
        return {
            unsubscribe() {
                this._subscriptions.set(key,
                    this._subscriptions.get(key).filter((cb) => cb !== callback));
            }
        };
    }

    onAllChanges(callback: StoreAllChangeSubscriptionFunction<T>): Subscription {
        this._allChangesSubscriptions.push(callback);
        return {
            unsubscribe() {
                this._allChangesSubscriptions = this._allChangesSubscriptions.filter((cb) => cb !== callback);
            }
        }
    }
}




