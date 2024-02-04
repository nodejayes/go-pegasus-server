interface Subscription {
  unsubscribe: () => void;
}

class MessageEventEmitter<T> {
  private _events: { [key: string]: { [key: string]: (payload: T) => void } } =
    {};

  subscribe(event: string, handler: (payload: any) => void): Subscription {
    let idx = 0;
    if (!this._events[event]) {
      this._events[event] = { [idx]: handler };
      return {
        unsubscribe: () => delete this._events[event][idx],
      };
    }
    idx = Object.keys(this._events[event]).length;
    this._events[event][idx] = handler;
    return {
      unsubscribe: () => delete this._events[event][idx],
    };
  }
  emit(event: string, payload?: T) {
    if (!this._events[event]) {
      return;
    }
    for (const evKey of Object.keys(this._events[event])) {
      const ev = this._events[event][evKey];
      if (!ev || typeof ev !== "function") {
        continue;
      }
      ev(payload ?? (null as T));
    }
  }
}

export { MessageEventEmitter, Subscription };
