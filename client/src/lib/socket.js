/**
 * Nihan WebSocket Client
 *
 * Uses native WebSocket (gorilla/websocket backend).
 * Auto-reconnects with exponential back-off.
 * Dispatches events via an internal EventEmitter pattern.
 */

const MAX_RECONNECT_DELAY = 30000;
const INITIAL_RECONNECT_DELAY = 1000;
const HEARTBEAT_INTERVAL = 25000;

class NihanSocket {
  constructor() {
    /** @type {WebSocket|null} */
    this._ws = null;
    /** @type {Map<string, Set<Function>>} */
    this._listeners = new Map();
    this._reconnectDelay = INITIAL_RECONNECT_DELAY;
    this._reconnectTimer = null;
    this._heartbeatTimer = null;
    this._intentionallyClosed = false;
    this._token = null;
    this._connected = false;
  }

  /* ───────── public API ───────── */

  /**
   * Connect to the WebSocket server.
   * @param {string} token - JWT auth token
   */
  connect(token) {
    this._token = token;
    this._intentionallyClosed = false;
    this._createConnection();
  }

  /**
   * Disconnect cleanly.
   */
  disconnect() {
    this._intentionallyClosed = true;
    this._clearTimers();
    if (this._ws) {
      this._ws.close(1000, 'Client disconnect');
      this._ws = null;
    }
    this._connected = false;
  }

  /**
   * Send a JSON message.
   * @param {string} type
   * @param {object} payload
   */
  send(type, payload = {}) {
    if (!this._ws || this._ws.readyState !== WebSocket.OPEN) {
      console.warn('[NihanSocket] Cannot send, not connected');
      return false;
    }
    this._ws.send(JSON.stringify({ type, ...payload }));
    return true;
  }

  /**
   * Send a chat message.
   */
  sendMessage(recipientId, ciphertext, nonce, messageId) {
    return this.send('message', {
      recipient_id: recipientId,
      ciphertext,
      nonce,
      message_id: messageId,
    });
  }

  /**
   * Send typing indicator.
   */
  sendTyping(recipientId) {
    return this.send('typing', { recipient_id: recipientId });
  }

  /**
   * Send read receipt.
   */
  sendReadReceipt(senderId, messageIds) {
    return this.send('read_receipt', {
      sender_id: senderId,
      message_ids: messageIds,
    });
  }

  /**
   * Check connection status.
   */
  get connected() {
    return this._connected;
  }

  /* ───────── event emitter ───────── */

  /**
   * Subscribe to an event.
   * @param {string} event
   * @param {Function} callback
   * @returns {Function} unsubscribe function
   */
  on(event, callback) {
    if (!this._listeners.has(event)) {
      this._listeners.set(event, new Set());
    }
    this._listeners.get(event).add(callback);
    return () => this.off(event, callback);
  }

  /**
   * Unsubscribe from an event.
   */
  off(event, callback) {
    const set = this._listeners.get(event);
    if (set) set.delete(callback);
  }

  /**
   * Emit an event to all listeners.
   */
  _emit(event, data) {
    const set = this._listeners.get(event);
    if (set) {
      set.forEach((cb) => {
        try {
          cb(data);
        } catch (err) {
          console.error(`[NihanSocket] Listener error for "${event}":`, err);
        }
      });
    }
  }

  /* ───────── internal ───────── */

  _createConnection() {
    if (this._ws && (this._ws.readyState === WebSocket.OPEN || this._ws.readyState === WebSocket.CONNECTING)) {
      return;
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    const url = `${protocol}//${host}/ws?token=${encodeURIComponent(this._token)}`;

    this._ws = new WebSocket(url);

    this._ws.onopen = () => {
      this._connected = true;
      this._reconnectDelay = INITIAL_RECONNECT_DELAY;
      this._startHeartbeat();
      this._emit('connected', null);
      console.log('[NihanSocket] Connected');
    };

    this._ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        const { type, ...payload } = data;

        if (type === 'pong') return; // heartbeat response

        this._emit(type, payload);
        this._emit('*', data); // wildcard
      } catch (err) {
        console.error('[NihanSocket] Failed to parse message:', err);
      }
    };

    this._ws.onclose = (event) => {
      this._connected = false;
      this._clearTimers();
      this._emit('disconnected', { code: event.code, reason: event.reason });
      console.log(`[NihanSocket] Disconnected (${event.code})`);

      if (!this._intentionallyClosed) {
        this._scheduleReconnect();
      }
    };

    this._ws.onerror = (error) => {
      this._emit('error', error);
      console.error('[NihanSocket] Error:', error);
    };
  }

  _startHeartbeat() {
    this._clearHeartbeat();
    this._heartbeatTimer = setInterval(() => {
      this.send('ping');
    }, HEARTBEAT_INTERVAL);
  }

  _clearHeartbeat() {
    if (this._heartbeatTimer) {
      clearInterval(this._heartbeatTimer);
      this._heartbeatTimer = null;
    }
  }

  _scheduleReconnect() {
    if (this._reconnectTimer) return;

    console.log(`[NihanSocket] Reconnecting in ${this._reconnectDelay}ms…`);
    this._emit('reconnecting', { delay: this._reconnectDelay });

    this._reconnectTimer = setTimeout(() => {
      this._reconnectTimer = null;
      this._createConnection();
    }, this._reconnectDelay);

    // Exponential back-off with jitter
    this._reconnectDelay = Math.min(
      this._reconnectDelay * 2 + Math.random() * 500,
      MAX_RECONNECT_DELAY
    );
  }

  _clearTimers() {
    this._clearHeartbeat();
    if (this._reconnectTimer) {
      clearTimeout(this._reconnectTimer);
      this._reconnectTimer = null;
    }
  }
}

// Singleton instance
const nihanSocket = new NihanSocket();

export default nihanSocket;
