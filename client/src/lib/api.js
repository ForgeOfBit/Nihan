/**
 * Nihan REST API Client
 *
 * All calls go through the Vite proxy (/api → http://localhost:8080/api).
 * Tokens are read from the authStore.
 */

const BASE = '/api';

/* ──────────────── internal helpers ──────────────── */

function getToken() {
  try {
    const raw = localStorage.getItem('nihan-auth');
    if (!raw) return null;
    const data = JSON.parse(raw);
    return data?.state?.token ?? null;
  } catch {
    return null;
  }
}

async function request(method, path, body = null, extraHeaders = {}) {
  const headers = { 'Content-Type': 'application/json', ...extraHeaders };
  const token = getToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;

  const options = { method, headers };
  if (body) options.body = JSON.stringify(body);

  const res = await fetch(`${BASE}${path}`, options);

  /* Handle 204 No Content */
  if (res.status === 204) return null;

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    const message = data?.error || data?.message || `Request failed (${res.status})`;
    const err = new Error(message);
    err.status = res.status;
    err.data = data;
    throw err;
  }

  return data;
}

/* ──────────────── Auth ──────────────── */

export const auth = {
  /**
   * Register a new user.
   * @param {{ username: string, email: string, password: string }} payload
   */
  register(payload) {
    return request('POST', '/auth/register', payload);
  },

  /**
   * Login with email & password.
   * @param {{ email: string, password: string }} payload
   */
  login(payload) {
    return request('POST', '/auth/login', payload);
  },

  /**
   * Refresh access token.
   * @param {{ refresh_token: string }} payload
   */
  refresh(payload) {
    return request('POST', '/auth/refresh', payload);
  },

  /**
   * Logout (invalidate token server-side).
   */
  logout() {
    return request('POST', '/auth/logout');
  },
};

/* ──────────────── Users ──────────────── */

export const users = {
  /**
   * Get current user profile.
   */
  me() {
    return request('GET', '/users/me');
  },

  /**
   * Update current user profile.
   * @param {{ display_name?: string, avatar_url?: string }} payload
   */
  updateProfile(payload) {
    return request('PATCH', '/users/me', payload);
  },

  /**
   * Search users by username#discriminator tag.
   * @param {string} tag - e.g. "alice#1234"
   */
  searchByTag(tag) {
    return request('GET', `/users/search?tag=${encodeURIComponent(tag)}`);
  },

  /**
   * Get a user's public profile by ID.
   * @param {string} userId
   */
  getById(userId) {
    return request('GET', `/users/${userId}`);
  },

  /**
   * Change discriminator (premium feature).
   * @param {{ discriminator: string }} payload
   */
  changeDiscriminator(payload) {
    return request('POST', '/users/me/discriminator', payload);
  },
};

/* ──────────────── Public Keys ──────────────── */

export const keys = {
  /**
   * Upload our public keys to the server.
   * @param {{ identity_key: string, exchange_key: string }} payload - Base64 encoded
   */
  upload(payload) {
    return request('POST', '/keys', payload);
  },

  /**
   * Get a user's public keys.
   * @param {string} userId
   */
  getByUser(userId) {
    return request('GET', `/keys/${userId}`);
  },
};

/* ──────────────── Friends ──────────────── */

export const friends = {
  /**
   * Get friend list.
   */
  list() {
    return request('GET', '/friends');
  },

  /**
   * Send a friend request.
   * @param {{ user_id?: string, tag?: string }} payload
   */
  sendRequest(payload) {
    return request('POST', '/friends/request', payload);
  },

  /**
   * Get pending friend requests.
   */
  pendingRequests() {
    return request('GET', '/friends/requests');
  },

  /**
   * Accept a friend request.
   * @param {string} requestId
   */
  accept(requestId) {
    return request('POST', `/friends/requests/${requestId}/accept`);
  },

  /**
   * Decline a friend request.
   * @param {string} requestId
   */
  decline(requestId) {
    return request('POST', `/friends/requests/${requestId}/decline`);
  },

  /**
   * Remove a friend.
   * @param {string} friendId
   */
  remove(friendId) {
    return request('DELETE', `/friends/${friendId}`);
  },
};

/* ──────────────── Messages ──────────────── */

export const messages = {
  /**
   * Get conversations list.
   */
  conversations() {
    return request('GET', '/messages/conversations');
  },

  /**
   * Get messages in a conversation.
   * @param {string} recipientId
   * @param {{ before?: string, limit?: number }} params
   */
  getMessages(recipientId, params = {}) {
    const qs = new URLSearchParams();
    if (params.before) qs.set('before', params.before);
    if (params.limit) qs.set('limit', String(params.limit));
    const query = qs.toString() ? `?${qs}` : '';
    return request('GET', `/messages/${recipientId}${query}`);
  },

  /**
   * Send a message (REST fallback — prefer WebSocket).
   * @param {{ recipient_id: string, ciphertext: string, nonce: string }} payload
   */
  send(payload) {
    return request('POST', '/messages', payload);
  },

  /**
   * Mark messages as read.
   * @param {string} conversationId
   */
  markRead(conversationId) {
    return request('POST', `/messages/${conversationId}/read`);
  },
};

export default { auth, users, keys, friends, messages };
