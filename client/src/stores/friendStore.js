/**
 * Friends Store — Zustand
 *
 * Manages friend list, requests, and user search.
 */

import { create } from 'zustand';
import { friends as friendsApi, users as usersApi } from '../lib/api.js';
import socket from '../lib/socket.js';

const useFriendStore = create((set, get) => ({
  /* ── state ── */
  friends: [],
  pendingRequests: [],
  searchResults: [],
  isLoadingFriends: false,
  isLoadingRequests: false,
  isSearching: false,
  searchQuery: '',
  error: null,
  wsListenersAttached: false,

  /* ── actions ── */

  /**
   * Load friends list.
   */
  loadFriends: async () => {
    set({ isLoadingFriends: true });
    try {
      const data = await friendsApi.list();
      set({ friends: data || [], isLoadingFriends: false });
    } catch (err) {
      console.error('[FriendStore] Failed to load friends:', err);
      set({ isLoadingFriends: false, error: err.message });
    }
  },

  /**
   * Load pending friend requests.
   */
  loadPendingRequests: async () => {
    set({ isLoadingRequests: true });
    try {
      const data = await friendsApi.pendingRequests();
      set({ pendingRequests: data || [], isLoadingRequests: false });
    } catch (err) {
      console.error('[FriendStore] Failed to load requests:', err);
      set({ isLoadingRequests: false, error: err.message });
    }
  },

  /**
   * Send a friend request by tag (e.g. "alice#1234").
   */
  sendFriendRequest: async (tag) => {
    set({ error: null });
    try {
      await friendsApi.sendRequest({ tag });
      return true;
    } catch (err) {
      set({ error: err.message });
      throw err;
    }
  },

  /**
   * Accept a friend request.
   */
  acceptRequest: async (requestId) => {
    try {
      await friendsApi.accept(requestId);
      set((state) => ({
        pendingRequests: state.pendingRequests.filter((r) => r.id !== requestId),
      }));
      // Reload friends list to include the new friend
      await get().loadFriends();
    } catch (err) {
      console.error('[FriendStore] Failed to accept request:', err);
      throw err;
    }
  },

  /**
   * Decline a friend request.
   */
  declineRequest: async (requestId) => {
    try {
      await friendsApi.decline(requestId);
      set((state) => ({
        pendingRequests: state.pendingRequests.filter((r) => r.id !== requestId),
      }));
    } catch (err) {
      console.error('[FriendStore] Failed to decline request:', err);
      throw err;
    }
  },

  /**
   * Remove a friend.
   */
  removeFriend: async (friendId) => {
    try {
      await friendsApi.remove(friendId);
      set((state) => ({
        friends: state.friends.filter((f) => f.id !== friendId && f.user_id !== friendId),
      }));
    } catch (err) {
      console.error('[FriendStore] Failed to remove friend:', err);
      throw err;
    }
  },

  /**
   * Search users by tag.
   */
  searchUsers: async (query) => {
    if (!query || query.length < 2) {
      set({ searchResults: [], searchQuery: query });
      return;
    }
    set({ isSearching: true, searchQuery: query });
    try {
      const data = await usersApi.searchByTag(query);
      set({ searchResults: data ? (Array.isArray(data) ? data : [data]) : [], isSearching: false });
    } catch (err) {
      console.error('[FriendStore] Search failed:', err);
      set({ searchResults: [], isSearching: false });
    }
  },

  /**
   * Clear search results.
   */
  clearSearch: () => set({ searchResults: [], searchQuery: '' }),

  /**
   * Clear error.
   */
  clearError: () => set({ error: null }),

  /**
   * Attach WebSocket listeners for friend events.
   */
  attachWSListeners: () => {
    if (get().wsListenersAttached) return;

    socket.on('friend_request', (data) => {
      set((state) => ({
        pendingRequests: [...state.pendingRequests, data],
      }));
    });

    socket.on('friend_accepted', (data) => {
      get().loadFriends();
      set((state) => ({
        pendingRequests: state.pendingRequests.filter(
          (r) => r.id !== data.request_id
        ),
      }));
    });

    socket.on('friend_removed', (data) => {
      set((state) => ({
        friends: state.friends.filter(
          (f) => f.id !== data.user_id && f.user_id !== data.user_id
        ),
      }));
    });

    set({ wsListenersAttached: true });
  },
}));

export default useFriendStore;
