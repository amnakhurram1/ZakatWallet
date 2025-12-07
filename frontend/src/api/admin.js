import { post } from './client.js';

/**
 * Fund a wallet via the admin faucet.  Creates a coinbase
 * transaction awarding tokens to the specified wallet.  The
 * amount recorded in the database may differ from the fixed
 * reward minted on chain.
 *
 * @param {{ address: string, amount: number }} data
 * @returns {Promise<{ address: string, amount: number, block_hash: string }>}
 */
export function fundWallet({ address, amount }) {
  if (!address || amount == null || amount <= 0) {
    return Promise.reject({ error: 'Invalid address or amount' });
  }
  return post('/admin/fund', { address, amount });
}
