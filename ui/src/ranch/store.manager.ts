import {CreateStore, Store} from "./store";

export interface StoreManager {
    CreateStore<T>(key: string): Store<T>;
    GetStore<T>(key: string): Store<T>;
    ResetStores(): void;
}

class storeManager implements StoreManager {
    private _stores: Map<string, Store<any>>;

    constructor() {
        this._stores = new Map<string, Store<any>>();
    }
    CreateStore<T>(key: string): Store<T> {
        const store: Store<T> = CreateStore<T>();
        this._stores.set(key, store);
        return store;
    }

    GetStore<T>(key: string): Store<T> {
        if (this._stores.has(key)) {
            return this._stores.get(key);
        }
        return CreateStore<T>();
    }

    ResetStores() {
        this._stores.forEach((store: Store<any>) => {
            store.reset();
        });
    }
}

let _storeManagerSingleton: StoreManager;
export function CreateStoreManager(): StoreManager {
    if (!_storeManagerSingleton) {
        _storeManagerSingleton = new storeManager();
    }
    return _storeManagerSingleton;
}

export function GetStoreManager(): StoreManager {
   return CreateStoreManager();
}