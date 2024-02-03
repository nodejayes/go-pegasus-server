import { MessageEventEmitter } from "./event.emitter";

export interface Message {
  type: string;
  payload: any;
}

export interface ActionResponse {
  code: number;
  error: string;
}

export interface EventHandlerConfig {
  eventUrl: string;
  actionUrl: string;
  clientIdHeaderKey: string;
  reconnectTimeout?: number;
}

class EventHandler {
  private _config: EventHandlerConfig | null = null;
  private _source: EventSource | null = null;
  private _sourceCanReconnect = true;
  private _sourceMessage = new MessageEventEmitter();
  private _readyConnection = new MessageEventEmitter();

  open(config: EventHandlerConfig) {
    if (!config.reconnectTimeout) {
      config.reconnectTimeout = 5000;
    }
    if (this._source) {
      this._sourceCanReconnect = false;
      this._source.close();
      this._sourceCanReconnect = true;
      this._source = null;
    }
    this._source = new EventSource(config.eventUrl);
    this._source.onmessage = (event: MessageEvent<any>) =>
      this.newMessage(event);
    this._source.onerror = (event: Event) => this.sourceError(event);
    this._source.onopen = (event) => this.sourceOpen(event);
    this._config = config;
  }

  listenEvent<T>(event: string, handler: (payload: T) => void) {
    this._sourceMessage.remove(event, handler);
    this._sourceMessage.on(event, handler);
  }

  async sendAction(message: Message): Promise<ActionResponse> {
    return await fetch(this._config.actionUrl, {
      mode: "cors",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(message),
    }).then((resp) => resp.json());
  }

  private newMessage(event: MessageEvent<any>) {
    const message = JSON.parse(event.data) as Message;
    this._sourceMessage.emit(message.type, message.payload);
  }

  private sourceError(event: any) {
    if (
      event?.target?.readyState === EventSource.CLOSED &&
      this._sourceCanReconnect &&
      this._config
    ) {
      setTimeout(() => this.open(this._config), this._config.reconnectTimeout);
    }
  }

  private sourceOpen(event: Event) {
    this._readyConnection.emit("ready");
  }
}

export const ServerEventHandler = new EventHandler();