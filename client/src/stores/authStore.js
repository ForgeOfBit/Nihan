/**
 * Auth Store — Zustand
 *
 * Manages user authentication state, tokens, and session lifecycle.
 * Persists to localStorage under "nihan-auth".
 */

import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import { auth as authApi, users as usersApi, keys as keysApi } from '../lib/api.js';
import {
  initCrypto,
  generateIdentityKeyPair,
  generateX25519KeyPair,
  exportPublicKeys,
  clearAllKeys,
} from '../lib/crypto.js';
import socket from '../lib/socket.js';

const useAuthStore = create(
  persist(
    (set, get) => ({
      /* ── state ── */
      user: null,
      token: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
      cryptoReady: false,

      /* ── actions ── */

      /**
       * Register a new account.
       */
      register: async ({ username, email, password }) => {
        set({ isLoading: true, error: null });
        try {
          const data = await authApi.register({ username, email, password });

          set({
            user: data.user,
            token: data.token,
            refreshToken: data.refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });

          // Generate E2EE keys and upload public parts
          await get().initializeKeys();

          // Connect WebSocket
          socket.connect(data.token);

          return data;
        } catch (err) {
          set({ isLoading: false, error: err.message });
          throw err;
        }
      },

      /**
       * Login with email + password.
       */
      login: async ({ email, password }) => {
        set({ isLoading: true, error: null });
        try {
          const data = await authApi.login({ email, password });

          set({
            user: data.user,
            token: data.token,
            refreshToken: data.refresh_token,
            isAuthenticated: true,
            isLoading: false,
          });

          // Init crypto and connect
          await get().initializeKeys();
          socket.connect(data.token);

          return data;
        } catch (err) {
          set({ isLoading: false, error: err.message });
          throw err;
        }
      },

      /**
       * Logout — clear everything.
       */
      logout: async () => {
        try {
          await authApi.logout();
        } catch {
          // ignore — we're logging out anyway
        }
        socket.disconnect();
        set({
          user: null,
          token: null,
          refreshToken: null,
          isAuthenticated: false,
          cryptoReady: false,
          error: null,
        });
      },

      /**
       * Initialize E2EE keys (generate if none exist, upload public keys).
       */
      initializeKeys: async () => {
        try {
          await initCrypto();
          const pubKeys = await exportPublicKeys();

          // Generate keys if not present
          if (!pubKeys.identityKey) {
            await generateIdentityKeyPair();
          }
          if (!pubKeys.exchangeKey) {
            await generateX25519KeyPair();
          }

          const finalKeys = await exportPublicKeys();

          // Upload public keys to the server
          try {
            await keysApi.upload({
              identity_key: finalKeys.identityKey,
              exchange_key: finalKeys.exchangeKey,
            });
          } catch {
            // keys might already be uploaded
          }

          set({ cryptoReady: true });
        } catch (err) {
          console.error('[AuthStore] Failed to initialize keys:', err);
        }
      },

      /**
       * Refresh user profile from server.
       */
      refreshProfile: async () => {
        try {
          const user = await usersApi.me();
          set({ user });
        } catch (err) {
          console.error('[AuthStore] Failed to refresh profile:', err);
        }
      },

      /**
       * Rehydrate session (called on app start).
       */
      rehydrate: async () => {
        const { token, isAuthenticated } = get();
        if (!token || !isAuthenticated) return;

        set({ isLoading: true });
        try {
          await initCrypto();
          set({ cryptoReady: true });

          const user = await usersApi.me();
          set({ user, isLoading: false });

          socket.connect(token);
        } catch (err) {
          // Token might be expired — try to refresh
          const { refreshToken } = get();
          if (refreshToken) {
            try {
              const data = await authApi.refresh({ refresh_token: refreshToken });
              set({
                token: data.token,
                refreshToken: data.refresh_token || refreshToken,
              });
              const user = await usersApi.me();
              set({ user, isLoading: false });
              socket.connect(data.token);
              return;
            } catch {
              // Refresh also failed
            }
          }

          // Clear session
          set({
            user: null,
            token: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
          });
        }
      },

      /**
       * Clear error message.
       */
      clearError: () => set({ error: null }),

      /**
       * Reset all keys (danger zone).
       */
      resetKeys: async () => {
        await clearAllKeys();
        await generateIdentityKeyPair();
        await generateX25519KeyPair();
        const finalKeys = await exportPublicKeys();
        await keysApi.upload({
          identity_key: finalKeys.identityKey,
          exchange_key: finalKeys.exchangeKey,
        });
        set({ cryptoReady: true });
      },
    }),
    {
      name: 'nihan-auth',
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

export default useAuthStore;
