/**
 * Chat Store — Zustand
 *
 * Manages conversations, messages, active chat, and real-time events.
 */

import { create } from 'zustand';
import { messages as messagesApi } from '../lib/api.js';
import socket from '../lib/socket.js';
import {
  encryptMessage,
  decryptMessage,
  getSessionKey,
  establishSession,
  fromBase64,
} from '../lib/crypto.js';
import { keys as keysApi } from '../lib/api.js';

const useChatStore = create((set, get) => ({
  /* ── state ── */
  conversations: [],
  messages: {}, // { [recipientId]: Message[] }
  activeChat: null, // recipientId
  activeChatUser: null, // user object
  typingUsers: {}, // { [userId]: timestamp }
  isLoadingConversations: false,
  isLoadingMessages: false,
  hasMoreMessages: {},
  wsListenersAttached: false,

  /* ── actions ── */

  /**
   * Load conversations list from the server.
   */
  loadConversations: async () => {
    set({ isLoadingConversations: true });
    try {
      const data = await messagesApi.conversations();
      set({ conversations: data || [], isLoadingConversations: false });
    } catch (err) {
      console.error('[ChatStore] Failed to load conversations:', err);
      set({ isLoadingConversations: false });
    }
  },

  /**
   * Set the active chat recipient.
   */
  setActiveChat: async (recipientId, user = null) => {
    set({
      activeChat: recipientId,
      activeChatUser: user,
    });

    if (recipientId) {
      await get().loadMessages(recipientId);
      // Mark as read
      try {
        await messagesApi.markRead(recipientId);
        // Update unread count in conversations
        set((state) => ({
          conversations: state.conversations.map((c) =>
            c.recipient_id === recipientId || c.user_id === recipientId
              ? { ...c, unread_count: 0 }
              : c
          ),
        }));
      } catch {
        // ignore
      }
    }
  },

  /**
   * Load messages for a conversation.
   */
  loadMessages: async (recipientId, before = null) => {
    set({ isLoadingMessages: true });
    try {
      const data = await messagesApi.getMessages(recipientId, { before, limit: 50 });
      const msgs = data || [];

      set((state) => {
        const existing = before ? (state.messages[recipientId] || []) : [];
        const merged = before ? [...msgs, ...existing] : msgs;

        return {
          messages: { ...state.messages, [recipientId]: merged },
          hasMoreMessages: { ...state.hasMoreMessages, [recipientId]: msgs.length >= 50 },
          isLoadingMessages: false,
        };
      });
    } catch (err) {
      console.error('[ChatStore] Failed to load messages:', err);
      set({ isLoadingMessages: false });
    }
  },

  /**
   * Send an encrypted message.
   */
  sendMessage: async (recipientId, plaintext) => {
    try {
      // Get or establish session key
      let sessionKey = await getSessionKey(recipientId);
      if (!sessionKey) {
        const peerKeys = await keysApi.getByUser(recipientId);
        if (peerKeys?.exchange_key) {
          sessionKey = await establishSession(recipientId, fromBase64(peerKeys.exchange_key));
        } else {
          throw new Error('Cannot establish secure session: peer has no exchange key');
        }
      }

      // Encrypt
      const { ciphertext, nonce } = encryptMessage(plaintext, sessionKey);

      // Generate a temporary ID
      const tempId = `temp_${Date.now()}_${Math.random().toString(36).slice(2)}`;

      // Optimistic UI update
      const optimisticMsg = {
        id: tempId,
        sender_id: 'self',
        recipient_id: recipientId,
        content: plaintext,
        ciphertext,
        nonce,
        created_at: new Date().toISOString(),
        status: 'sending',
        encrypted: true,
      };

      set((state) => ({
        messages: {
          ...state.messages,
          [recipientId]: [...(state.messages[recipientId] || []), optimisticMsg],
        },
      }));

      // Send via WebSocket
      const sent = socket.sendMessage(recipientId, ciphertext, nonce, tempId);

      if (!sent) {
        // Fallback to REST
        await messagesApi.send({ recipient_id: recipientId, ciphertext, nonce });
      }

      // Mark as sent
      set((state) => ({
        messages: {
          ...state.messages,
          [recipientId]: (state.messages[recipientId] || []).map((m) =>
            m.id === tempId ? { ...m, status: 'sent' } : m
          ),
        },
      }));
    } catch (err) {
      console.error('[ChatStore] Failed to send message:', err);
      throw err;
    }
  },

  /**
   * Handle incoming message from WebSocket.
   */
  receiveMessage: async (data) => {
    const senderId = data.sender_id;
    let content = data.ciphertext;

    // Try to decrypt
    if (data.ciphertext && data.nonce) {
      try {
        const sessionKey = await getSessionKey(senderId);
        if (sessionKey) {
          content = decryptMessage(data.ciphertext, data.nonce, sessionKey);
        }
      } catch (err) {
        console.warn('[ChatStore] Failed to decrypt message:', err);
        content = '[Encrypted message — unable to decrypt]';
      }
    }

    const msg = {
      id: data.id || data.message_id,
      sender_id: senderId,
      recipient_id: data.recipient_id,
      content,
      ciphertext: data.ciphertext,
      nonce: data.nonce,
      created_at: data.created_at || new Date().toISOString(),
      status: 'received',
      encrypted: !!data.ciphertext,
    };

    set((state) => ({
      messages: {
        ...state.messages,
        [senderId]: [...(state.messages[senderId] || []), msg],
      },
    }));

    // Update conversations list
    get().updateConversationPreview(senderId, content);

    // If this chat is active, mark as read
    if (get().activeChat === senderId) {
      try {
        await messagesApi.markRead(senderId);
        socket.sendReadReceipt(senderId, [msg.id]);
      } catch {
        // ignore
      }
    }
  },

  /**
   * Update a conversation preview when a new message arrives.
   */
  updateConversationPreview: (userId, lastMessage) => {
    set((state) => {
      const existing = state.conversations.find(
        (c) => c.recipient_id === userId || c.user_id === userId
      );

      if (existing) {
        return {
          conversations: state.conversations.map((c) =>
            c.recipient_id === userId || c.user_id === userId
              ? {
                  ...c,
                  last_message: lastMessage,
                  last_message_at: new Date().toISOString(),
                  unread_count: state.activeChat === userId ? 0 : (c.unread_count || 0) + 1,
                }
              : c
          ),
        };
      }

      // Create new conversation entry
      return {
        conversations: [
          {
            user_id: userId,
            recipient_id: userId,
            last_message: lastMessage,
            last_message_at: new Date().toISOString(),
            unread_count: state.activeChat === userId ? 0 : 1,
          },
          ...state.conversations,
        ],
      };
    });
  },

  /**
   * Handle typing indicator.
   */
  setTyping: (userId) => {
    set((state) => ({
      typingUsers: { ...state.typingUsers, [userId]: Date.now() },
    }));

    // Clear after 3 seconds
    setTimeout(() => {
      set((state) => {
        const updated = { ...state.typingUsers };
        if (updated[userId] && Date.now() - updated[userId] >= 2800) {
          delete updated[userId];
        }
        return { typingUsers: updated };
      });
    }, 3000);
  },

  /**
   * Attach WebSocket listeners.
   */
  attachWSListeners: () => {
    if (get().wsListenersAttached) return;

    socket.on('message', (data) => {
      get().receiveMessage(data);
    });

    socket.on('typing', (data) => {
      get().setTyping(data.sender_id || data.user_id);
    });

    socket.on('read_receipt', (data) => {
      const senderId = data.sender_id || data.user_id;
      const messageIds = data.message_ids || [];
      set((state) => {
        const updated = { ...state.messages };
        Object.keys(updated).forEach((recipientId) => {
          updated[recipientId] = updated[recipientId].map((m) =>
            messageIds.includes(m.id) ? { ...m, status: 'read' } : m
          );
        });
        return { messages: updated };
      });
    });

    socket.on('message_ack', (data) => {
      const tempId = data.temp_id || data.message_id;
      const realId = data.id || data.real_id;
      set((state) => {
        const updated = { ...state.messages };
        Object.keys(updated).forEach((recipientId) => {
          updated[recipientId] = updated[recipientId].map((m) =>
            m.id === tempId ? { ...m, id: realId || m.id, status: 'sent' } : m
          );
        });
        return { messages: updated };
      });
    });

    set({ wsListenersAttached: true });
  },

  /**
   * Clear active chat.
   */
  clearActiveChat: () => {
    set({ activeChat: null, activeChatUser: null });
  },
}));

export default useChatStore;
