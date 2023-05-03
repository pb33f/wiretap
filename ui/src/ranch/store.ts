import {Subscription} from "./bus";
export type StoreValueSubscriptionFunction<T> = (value: T) => void;
export type StoreAllChangeSubscriptionFunction<T> = (key: string, value: T) => void;
export type StorePopulatedSubscriptionFunction<T> = (store: Map<string, T>) => void;


export interface Store<T> {
    set(key: string, value: T): void;
    get(key: string): T;
    populate(data: Map<string, T>): void
    export()
    subscribe(key: string, callback: StoreValueSubscriptionFunction<T>): Subscription;
    onAllChanges(callback: StoreAllChangeSubscriptionFunction<T>): Subscription;
    onPopulated(callback: StorePopulatedSubscriptionFunction<T>): Subscription;
}

export function CreateStore<T>(): Store<T> {
    return new store<T>();
}

class store<T> {
    private _values: Map<string, T>;
    private _subscriptions: Map<string, StoreValueSubscriptionFunction<T>[]>;
    private _allChangesSubscriptions: StoreAllChangeSubscriptionFunction<T>[];
    private _storePopulatedSubscriptions: StorePopulatedSubscriptionFunction<T>[];

    constructor() {
        this._values = new Map<string, T>();
        this._subscriptions = new Map<string, StoreValueSubscriptionFunction<T>[]>()
        this._allChangesSubscriptions = [];
        this._storePopulatedSubscriptions = [];
    }

    set(key: string, value: T): void {
        this._values.set(key, value);
        this.alertSubscribers(key, value)
    }

    private alertSubscribers(key: string, value: T): void {
        if (this._subscriptions.has(key)) {
            this._subscriptions.get(key).forEach(
                (callback: StoreValueSubscriptionFunction<T>) => callback(value));
        }
        if(this._allChangesSubscriptions.length > 0) {
            this._allChangesSubscriptions.forEach(
                (callback: StoreAllChangeSubscriptionFunction<T>) => callback(key, value));
        }
    }

    get(key: string): T {
        return this._values.get(key);
    }

    populate(data: Map<string, T>): void {
        if (data && data.size > 0) {
            this._values = data;
            if(this._storePopulatedSubscriptions.length > 0) {
                this._storePopulatedSubscriptions.forEach(
                    (callback: StorePopulatedSubscriptionFunction<T>) => callback(data));
            }
        }
    }

    export(): Map<string, T> {
        return this._values
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
                this._allChangesSubscriptions =
                    this._allChangesSubscriptions.filter((cb) => cb !== callback);
            }
        }
    }

    onPopulated(callback: StorePopulatedSubscriptionFunction<T>): Subscription {
        this._storePopulatedSubscriptions.push(callback);
        return {
            unsubscribe() {
                this._storePopulatedSubscriptions =
                    this._storePopulatedSubscriptions.filter((cb) => cb !== callback);
            }
        }
    }

}




