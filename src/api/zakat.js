import { post } from './client.js';

/**
 * Run the global zakat deduction.  This endpoint will iterate
 * over all wallet profiles and deduct the zakat amount from
 * each, mining a block for each deduction.
 *
 * @returns {Promise<{ total_wallets: number, processed: number, total_zakat: number, block_hashes: string[] }>}
 */
export function runZakat() {
  return post('/zakat/run', {});
}
