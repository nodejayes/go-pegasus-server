class MessageEventEmitter<T> {
  private _events: { [key: string]: ((payload: T) => void)[] } = {};

  on(event: string, handler: (payload: T) => void) {
    if (!this._events[event]) {
      this._events[event] = [handler];
      return;
    }
    this._events[event].push(handler);
  }
  remove(event: string, handlerToRemove: (payload: T) => void) {
    if (!this._events[event]) {
      return;
    }
    this._events[event] = this._events[event].filter((listener) => listener !== handlerToRemove);
  }
  emit(event: string, payload?: T) {
    if (!this._events[event]) {
      return;
    }
    for (const ev of this._events[event]) {
      if (typeof ev !== 'function') {
        continue;
      }
      ev(payload ?? null);
    }
  }
}

export {MessageEventEmitter}