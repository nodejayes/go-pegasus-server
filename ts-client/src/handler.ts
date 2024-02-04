import { MessageEventEmitter, Subscription } from "./event.emitter";
import { v4 } from "uuid";

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
      config.eventUrl = "/events";
      config.actionUrl = "/action";
      config.clientIdHeaderKey = "clientId";
      config.reconnectTimeout = 5000;
    }
    this._config = config;
    if (this._source) {
      this._sourceCanReconnect = false;
      this._source.close();
      this._sourceCanReconnect = true;
      this._source = null;
    }
    const eventUrl = `${
      config.eventUrl.endsWith("/")
        ? config.eventUrl.substring(0, config.eventUrl.length - 1)
        : config.eventUrl
    }?clientId=${this.getClientId(config?.clientIdHeaderKey ?? "")}`;
    this._source = new EventSource(eventUrl);
    this._source.onmessage = (event: MessageEvent<any>) =>
      this.newMessage(event);
    this._source.onerror = (event: Event) => this.sourceError(event);
    this._source.onopen = (event) => this.sourceOpen(event);
  }

  subscribe<T>(event: string, handler: (payload: T) => void): Subscription {
    return this._sourceMessage.subscribe(event, handler);
  }

  async sendAction(message: Message): Promise<ActionResponse> {
    if (!this._config) {
      throw new Error("no config found");
    }
    return await fetch(this._config.actionUrl, {
      mode: "cors",
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        [this._config.clientIdHeaderKey]: this.getClientId(
          this._config.clientIdHeaderKey
        ),
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
      setTimeout(
        () =>
          this._config ? this.open(this._config) : this.sourceError(event),
        this._config.reconnectTimeout
      );
    }
  }

  private sourceOpen(event: Event) {
    this._readyConnection.emit("ready");
  }

  private getClientId(key: string): string {
    let clientId = localStorage.getItem(key);
    if (!clientId) {
      clientId = v4();
      localStorage.setItem(key, clientId);
    }
    return clientId;
  }
}

export const ServerEventHandler = new EventHandler();
