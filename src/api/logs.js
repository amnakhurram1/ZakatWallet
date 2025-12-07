import { get } from './client.js';

/**
 * Retrieve recent system log entries.  Limit can be specified
 * between 1 and 1000; defaults to 100.
 *
 * @param {number} [limit=100] Number of log entries to return
 * @returns {Promise<{ logs: Array<{ id: string, level: string, type: string, message: string, ip: string, timestamp: string }> }>}
 */
export function getSystemLogs(limit = 100) {
  const l = Math.min(Math.max(parseInt(limit, 10) || 100, 1), 1000);
  return get(`/logs/system?limit=${l}`);
}
