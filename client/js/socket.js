/**
 * A wrapper around a WebSocket connection.
 */
class Socket extends EventTarget {
    ackTimeoutMS = 5000;
    open = false;
    callbacks = {};

    constructor(hostname) {
        super();
        this.s = new WebSocket(`ws://${hostname}`)
        this.s.addEventListener('open', this.onOpen.bind(this));
        this.s.addEventListener('message', this.onMessage.bind(this));
        this.s.addEventListener('close', this.onClose.bind(this));
        this.s.addEventListener('error', this.onError.bind(this));
    }

    onOpen(event) {
        this.open = true;
        this.dispatchEvent(new Event("open"));
        console.log("websocket opened:", event);
    }

    onMessage(event) {
        let data = JSON.parse(event.data);
        if (data.id) {
            this.dispatchEvent(new CustomEvent(data.id, { detail: data }))
            return;
        }
        this.dispatch(data);
    }

    onClose(event) {
        this.open = false;
        this.dispatchEvent(new Event("closed"));
        console.log("websocket closed:", event);
    }

    onError(event) {
        this.dispatchEvent(new CustomEvent("error", { detail: event }));
        console.error("websocket error:", event);
    }

    send(obj) {
        if (!open) {
            console.log("send attempted before websocket connection was open");
            return false;
        }
        this.s.send(JSON.stringify(obj));
        return true;
    }

    sendWithResponse(obj) {
        return new Promise((resolve, reject) => {
            if (!obj.id) {
                reject("input must be an object with an \"id\" property");
                return;
            }
            if (!this.send(obj)) {
                reject("sendWithResponse attempted before websocket connection was open");
                return;
            }
            this.addEventListener(obj.id, o => resolve(o.detail));
            setTimeout(() => reject(`response ${obj.id} not received after ${this.ackTimeoutMS}ms`), this.ackTimeoutMS);
        });
    }

    dispatch(data) {
        if (!data.type) {
            console.error("received message with no type:", data);
            return;
        }
        let callback = this.callbacks[data.type];
        if (!callback) {
            console.log("received message with unregistered type:", data);
            return;
        }
        callback(data);
    }

    register(type, cb) {
        this.callbacks[type] = cb;
    }
}
