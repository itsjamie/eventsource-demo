interface Window {
    EventSource: sse.IEventSource
}

declare module sse {
    enum ReadyState {
        CONNECTING = 0, 
        OPEN = 1, 
        CLOSED = 2,
    }
    
    interface IEventSource extends EventTarget {
        new (url: string): IEventSource;
        new (url: string, dict: INewOptions): IEventSource;
        url: string;
        withCredentials: boolean;
        CONNECTING: ReadyState; // constant, always 0
        OPEN: ReadyState; // constant, always 1
        CLOSED: ReadyState; // constant, always 2
        readyState: ReadyState;
        onopen: Function;
        onmessage: (e: MessageEvent) => void;
        onerror: ErrorEventHandler;
        close: () => void;
    }
    
    interface INewOptions {
        withCredentials?: boolean;
    }
}