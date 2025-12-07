import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { getBlockByIndex } from "../api/blocks.js";
import { FiBox, FiHash, FiDatabase } from "react-icons/fi";

export default function BlockDetailPage() {
  const { index } = useParams();
  const [block, setBlock] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (index === undefined) return;
    setLoading(true);
    setError(null);

    getBlockByIndex(index)
      .then((blk) => setBlock(blk))
      .catch((err) =>
        setError(err?.message || err?.error || String(err))
      )
      .finally(() => setLoading(false));
  }, [index]);

  return (
    <div className="space-y-8">

      {/* TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiBox size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">
          Block #{index}
        </h2>
      </div>

      {/* LOADING */}
      {loading && (
        <p className="text-gray-300 animate-pulse">Loading block…</p>
      )}

      {/* ERROR */}
      {error && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded border border-red-700">
          {error}
        </p>
      )}

      {/* BLOCK CONTENT */}
      {!loading && !error && block && (
        <div className="space-y-10">

          {/* METADATA PANEL */}
          <div className="glass p-6 rounded-2xl shadow-xl space-y-4">

            <div className="flex items-center gap-3">
              <FiHash className="text-secondary" size={28} />
              <h3 className="text-xl font-semibold text-gray-200">
                Block Metadata
              </h3>
            </div>

            <div className="space-y-3 text-gray-300">
              <p>
                <span className="font-semibold text-primary">Timestamp:</span>{" "}
                {block.Timestamp
                  ? new Date(block.Timestamp * 1000).toLocaleString()
                  : ""}
              </p>

              <p className="break-all font-mono">
                <span className="font-semibold text-primary">Hash:</span>{" "}
                {block.Hash}
              </p>

              <p className="break-all font-mono">
                <span className="font-semibold text-primary">Previous Hash:</span>{" "}
                {block.PrevHash}
              </p>

              <p>
                <span className="font-semibold text-primary">Nonce:</span>{" "}
                {block.Nonce}
              </p>
            </div>
          </div>

          {/* TRANSACTIONS PANEL */}
          <div className="glass p-6 rounded-2xl shadow-xl">

            <div className="flex items-center gap-3 mb-4">
              <FiDatabase className="text-secondary" size={28} />
              <h3 className="text-xl font-semibold text-gray-200">
                Transactions ({block.Transactions.length})
              </h3>
            </div>

            {block.Transactions.length === 0 ? (
              <p className="text-gray-400">No transactions in this block.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full border-collapse">
                  <thead>
                    <tr className="bg-gradient-to-r from-primary to-secondary text-black">
                      <th className="py-3 px-4 text-left text-sm font-semibold">
                        Tx ID
                      </th>
                      <th className="py-3 px-4 text-left text-sm font-semibold">
                        Outputs
                      </th>
                      <th className="py-3 px-4 text-left text-sm font-semibold">
                        Total Value
                      </th>
                    </tr>
                  </thead>

                  <tbody>
                    {block.Transactions.map((tx) => {
                      const totalValue = Array.isArray(tx.Vout)
                        ? tx.Vout.reduce(
                            (sum, o) => sum + (o.Value || 0),
                            0
                          )
                        : 0;

                      return (
                        <tr
                          key={tx.ID}
                          className="border-t border-gray-700 hover:bg-white/5 transition-all"
                        >
                          <td className="py-3 px-4 text-gray-300 font-mono text-xs break-all">
                            {tx.ID}
                          </td>

                          <td className="py-3 px-4 text-gray-300">
                            {tx.Vout?.length || 0}
                          </td>

                          <td className="py-3 px-4 text-lg font-bold text-primary">
                            {totalValue}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
