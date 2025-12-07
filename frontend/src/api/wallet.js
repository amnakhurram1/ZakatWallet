import { get } from './client.js';

/**
 * Fetch the current balance for a wallet.
 *
 * @param {string} address Wallet address (hex string)
 * @returns {Promise<{ balance: number }>}
 */
export function getBalance(address) {
  if (!address) {
    return Promise.reject({ error: 'Address is required' });
  }
  return get(`/wallets/${address}/balance`);
}

/**
 * Fetch a full report for a wallet, including total sent/received
 * amounts and zakat records.
 *
 * @param {string} address Wallet address (hex string)
 * @returns {Promise<{ wallet_address: string, balance: number, total_sent: number, total_received: number, total_zakat: number, transactions: Array, zakat_records: Array }>}
 */
export function getWalletReport(address) {
  if (!address) {
    return Promise.reject({ error: 'Address is required' });
  }
  return get(`/reports/wallet/${address}`);
}
