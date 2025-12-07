import { get } from './client.js';

/**
 * Fetch summaries of all blocks in the blockchain.
 *
 * @returns {Promise<Array<{ index: number, timestamp: number, hash: string, prev_hash: string, tx_count: number }>>}
 */
export function getBlocks() {
  return get('/blocks');
}

/**
 * Fetch the full details of a specific block by its index.
 *
 * @param {number|string} index Zero‑based block height
 * @returns {Promise<{ Timestamp: number, Transactions: Array, PrevHash: string, Hash: string, Nonce: number }>}
 */
export function getBlockByIndex(index) {
  if (index === undefined || index === null) {
    return Promise.reject({ error: 'Index is required' });
  }
  return get(`/blocks/${index}`);
}
