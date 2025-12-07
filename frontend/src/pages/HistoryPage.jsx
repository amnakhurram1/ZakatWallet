import React, { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import { getWalletReport } from "../api/wallet.js";
import { FiClock, FiArrowUpRight, FiArrowDownLeft } from "react-icons/fi";

export default function HistoryPage() {
  const { walletAddress } = useAuth();
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!walletAddress) return;
    setLoading(true);
    setError(null);

    getWalletReport(walletAddress)
      .then((report) => {
        setTransactions(Array.isArray(report?.transactions) ? report.transactions : []);
      })
      .catch((err) => {
        setError(err?.message || err?.error || String(err));
      })
      .finally(() => setLoading(false));
  }, [walletAddress]);

  return (
    <div className="space-y-8">

      {/* TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiClock size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">Transaction History</h2>
      </div>

      {/* LOADING */}
      {loading && (
        <p className="text-gray-300 animate-pulse text-center">Loading…</p>
      )}

      {/* ERROR */}
      {error && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded text-center border border-red-700">
          {error}
        </p>
      )}

      {/* NO DATA */}
      {!loading && !error && transactions.length === 0 && (
        <p className="text-gray-400 text-center">No transactions found.</p>
      )}

      {/* TRANSACTION TABLE */}
      {!loading && !error && transactions.length > 0 && (
        <div className="glass rounded-2xl shadow-xl overflow-hidden">

          <table className="w-full border-collapse">
            <thead>
              <tr className="bg-gradient-to-r from-primary to-secondary text-black">
                <th className="py-3 px-4 text-left text-sm font-semibold">Date</th>
                <th className="py-3 px-4 text-left text-sm font-semibold">Type</th>
                <th className="py-3 px-4 text-left text-sm font-semibold">Sender</th>
                <th className="py-3 px-4 text-left text-sm font-semibold">Receiver</th>
                <th className="py-3 px-4 text-left text-sm font-semibold">Amount</th>
              </tr>
            </thead>

            <tbody>
              {transactions.map((tx) => {
                const date = tx.timestamp
                  ? new Date(tx.timestamp).toLocaleString()
                  : "";

                const isSent = tx.sender?.toLowerCase() === walletAddress?.toLowerCase();

                return (
                  <tr
                    key={tx.txid}
                    className="border-t border-gray-700 hover:bg-white/5 transition-all"
                  >
                    {/* DATE */}
                    <td className="py-3 px-4 text-gray-300">{date}</td>

                    {/* TYPE WITH ICON */}
                    <td className="py-3 px-4 text-gray-300 flex items-center gap-2">
                      {isSent ? (
                        <FiArrowUpRight className="text-red-400" />
                      ) : (
                        <FiArrowDownLeft className="text-green-400" />
                      )}
                      <span className={isSent ? "text-red-400" : "text-green-400"}>
                        {tx.type || (isSent ? "Sent" : "Received")}
                      </span>
                    </td>

                    {/* SENDER */}
                    <td className="py-3 px-4 text-gray-400 font-mono text-xs break-all">
                      {tx.sender}
                    </td>

                    {/* RECEIVER */}
                    <td className="py-3 px-4 text-gray-400 font-mono text-xs break-all">
                      {tx.receiver}
                    </td>

                    {/* AMOUNT */}
                    <td
                      className={`py-3 px-4 text-lg font-bold ${
                        isSent ? "text-red-400" : "text-green-400"
                      }`}
                    >
                      {tx.amount}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>

        </div>
      )}
    </div>
  );
}
