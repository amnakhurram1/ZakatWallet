import { post } from './client.js';

/**
 * Submit a new transaction to the backend.  The server will
 * construct and sign the transaction, mine it into a block and
 * update the UTXO set.
 *
 * @param {{ from: string, to: string, amount: number, privKey: string }} data
 * @returns {Promise<{ status: string }>}
 */
export function sendTransaction({ from, to, amount, privKey }) {
  if (!from || !to || amount == null || amount <= 0 || !privKey) {
    return Promise.reject({ error: 'Invalid transaction parameters' });
  }
  return post('/transactions', { from, to, amount, privKey });
}
