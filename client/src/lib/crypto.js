/**
 * Nihan E2EE Crypto Module
 *
 * Uses libsodium for:
 * - X25519 key exchange (Curve25519 Diffie-Hellman)
 * - XChaCha20-Poly1305 authenticated encryption
 * - Ed25519 digital signatures
 * - IndexedDB for persistent key storage via `idb`
 */

import sodium from 'libsodium-wrappers';
import { openDB } from 'idb';

const DB_NAME = 'nihan-keystore';
const DB_VERSION = 1;
const STORE_KEYS = 'keys';
const STORE_SESSION = 'sessions';

let _db = null;
let _ready = false;

/**
 * Initialize libsodium and open the IndexedDB key store.
 */
export async function initCrypto() {
  if (_ready) return;
  await sodium.ready;

  _db = await openDB(DB_NAME, DB_VERSION, {
    upgrade(db) {
      if (!db.objectStoreNames.contains(STORE_KEYS)) {
        db.createObjectStore(STORE_KEYS);
      }
      if (!db.objectStoreNames.contains(STORE_SESSION)) {
        db.createObjectStore(STORE_SESSION);
      }
    },
  });

  _ready = true;
}

function ensureReady() {
  if (!_ready) throw new Error('Crypto not initialized. Call initCrypto() first.');
}

/* ──────────────────── IndexedDB helpers ──────────────────── */

async function dbPut(store, key, value) {
  ensureReady();
  await _db.put(store, value, key);
}

async function dbGet(store, key) {
  ensureReady();
  return _db.get(store, key);
}

async function dbDelete(store, key) {
  ensureReady();
  await _db.delete(store, key);
}

async function dbClear(store) {
  ensureReady();
  await _db.clear(store);
}

/* ──────────────────── Identity key pair (Ed25519) ──────────────────── */

/**
 * Generate a new Ed25519 signing key pair and persist it.
 * @returns {{ publicKey: Uint8Array, privateKey: Uint8Array }}
 */
export async function generateIdentityKeyPair() {
  ensureReady();
  const kp = sodium.crypto_sign_keypair();
  await dbPut(STORE_KEYS, 'identity_public', kp.publicKey);
  await dbPut(STORE_KEYS, 'identity_private', kp.privateKey);
  return { publicKey: kp.publicKey, privateKey: kp.privateKey };
}

/**
 * Retrieve the stored identity key pair.
 */
export async function getIdentityKeyPair() {
  ensureReady();
  const publicKey = await dbGet(STORE_KEYS, 'identity_public');
  const privateKey = await dbGet(STORE_KEYS, 'identity_private');
  if (!publicKey || !privateKey) return null;
  return { publicKey, privateKey };
}

/* ──────────────────── X25519 key exchange ──────────────────── */

/**
 * Generate a new X25519 key pair for Diffie-Hellman key exchange and persist it.
 * @returns {{ publicKey: Uint8Array, privateKey: Uint8Array }}
 */
export async function generateX25519KeyPair() {
  ensureReady();
  const kp = sodium.crypto_box_keypair();
  await dbPut(STORE_KEYS, 'x25519_public', kp.publicKey);
  await dbPut(STORE_KEYS, 'x25519_private', kp.privateKey);
  return { publicKey: kp.publicKey, privateKey: kp.privateKey };
}

/**
 * Retrieve the stored X25519 key pair.
 */
export async function getX25519KeyPair() {
  ensureReady();
  const publicKey = await dbGet(STORE_KEYS, 'x25519_public');
  const privateKey = await dbGet(STORE_KEYS, 'x25519_private');
  if (!publicKey || !privateKey) return null;
  return { publicKey, privateKey };
}

/**
 * Derive a shared secret from our private key and the peer's public key.
 * Uses X25519 scalar multiplication → BLAKE2b hash for a 256-bit shared key.
 * @param {Uint8Array} ourPrivateKey
 * @param {Uint8Array} theirPublicKey
 * @returns {Uint8Array} 32-byte shared secret
 */
export function deriveSharedSecret(ourPrivateKey, theirPublicKey) {
  ensureReady();
  const raw = sodium.crypto_scalarmult(ourPrivateKey, theirPublicKey);
  return sodium.crypto_generichash(32, raw);
}

/* ──────────────────── Session key management ──────────────────── */

/**
 * Compute + cache a session key with a peer.
 * @param {string} peerId
 * @param {Uint8Array} theirPublicKey
 */
export async function establishSession(peerId, theirPublicKey) {
  ensureReady();
  const ourKeys = await getX25519KeyPair();
  if (!ourKeys) throw new Error('No X25519 key pair found. Generate keys first.');
  const shared = deriveSharedSecret(ourKeys.privateKey, theirPublicKey);
  await dbPut(STORE_SESSION, `session_${peerId}`, shared);
  return shared;
}

/**
 * Retrieve a cached session key.
 * @param {string} peerId
 * @returns {Uint8Array|null}
 */
export async function getSessionKey(peerId) {
  return dbGet(STORE_SESSION, `session_${peerId}`);
}

/* ──────────────────── Encryption (XChaCha20-Poly1305) ──────────────────── */

/**
 * Encrypt a plaintext message.
 * @param {string} plaintext
 * @param {Uint8Array} key - 32-byte symmetric key
 * @returns {{ ciphertext: string, nonce: string }} Base64-encoded
 */
export function encryptMessage(plaintext, key) {
  ensureReady();
  const nonce = sodium.randombytes_buf(sodium.crypto_aead_xchacha20poly1305_ietf_NPUBBYTES);
  const encoded = sodium.from_string(plaintext);
  const ciphertext = sodium.crypto_aead_xchacha20poly1305_ietf_encrypt(
    encoded,
    null, // additional data
    null, // nsec (unused)
    nonce,
    key
  );
  return {
    ciphertext: sodium.to_base64(ciphertext, sodium.base64_variants.ORIGINAL),
    nonce: sodium.to_base64(nonce, sodium.base64_variants.ORIGINAL),
  };
}

/**
 * Decrypt a ciphertext message.
 * @param {string} ciphertextB64
 * @param {string} nonceB64
 * @param {Uint8Array} key - 32-byte symmetric key
 * @returns {string} plaintext
 */
export function decryptMessage(ciphertextB64, nonceB64, key) {
  ensureReady();
  const ciphertext = sodium.from_base64(ciphertextB64, sodium.base64_variants.ORIGINAL);
  const nonce = sodium.from_base64(nonceB64, sodium.base64_variants.ORIGINAL);
  const plaintext = sodium.crypto_aead_xchacha20poly1305_ietf_decrypt(
    null, // nsec (unused)
    ciphertext,
    null, // additional data
    nonce,
    key
  );
  return sodium.to_string(plaintext);
}

/* ──────────────────── Signing (Ed25519) ──────────────────── */

/**
 * Sign a message with our Ed25519 private key.
 * @param {string} message
 * @param {Uint8Array} privateKey
 * @returns {string} Base64-encoded signature
 */
export function signMessage(message, privateKey) {
  ensureReady();
  const encoded = sodium.from_string(message);
  const signature = sodium.crypto_sign_detached(encoded, privateKey);
  return sodium.to_base64(signature, sodium.base64_variants.ORIGINAL);
}

/**
 * Verify a detached signature.
 * @param {string} message
 * @param {string} signatureB64
 * @param {Uint8Array} publicKey
 * @returns {boolean}
 */
export function verifySignature(message, signatureB64, publicKey) {
  ensureReady();
  const encoded = sodium.from_string(message);
  const signature = sodium.from_base64(signatureB64, sodium.base64_variants.ORIGINAL);
  return sodium.crypto_sign_verify_detached(signature, encoded, publicKey);
}

/* ──────────────────── Utility helpers ──────────────────── */

/**
 * Encode bytes to Base64.
 */
export function toBase64(bytes) {
  ensureReady();
  return sodium.to_base64(bytes, sodium.base64_variants.ORIGINAL);
}

/**
 * Decode Base64 to bytes.
 */
export function fromBase64(b64) {
  ensureReady();
  return sodium.from_base64(b64, sodium.base64_variants.ORIGINAL);
}

/**
 * Encode bytes to hex string.
 */
export function toHex(bytes) {
  ensureReady();
  return sodium.to_hex(bytes);
}

/**
 * Wipe all keys from IndexedDB.
 */
export async function clearAllKeys() {
  ensureReady();
  await dbClear(STORE_KEYS);
  await dbClear(STORE_SESSION);
}

/**
 * Export the public keys as Base64 for server registration.
 */
export async function exportPublicKeys() {
  ensureReady();
  const identity = await getIdentityKeyPair();
  const x25519 = await getX25519KeyPair();
  return {
    identityKey: identity ? toBase64(identity.publicKey) : null,
    exchangeKey: x25519 ? toBase64(x25519.publicKey) : null,
  };
}

export { sodium };
