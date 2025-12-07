import React, { useEffect, useState } from "react";
import { getBlocks } from "../api/blocks.js";
import { useNavigate } from "react-router-dom";
import { FiPackage } from "react-icons/fi";

export default function BlocksPage() {
  const [blocks, setBlocks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    setLoading(true);
    setError(null);

    getBlocks()
      .then((list) => setBlocks(Array.isArray(list) ? list : []))
      .catch((err) => setError(err?.message || err?.error || String(err)))
      .finally(() => setLoading(false));
  }, []);

  const handleRowClick = (index) => {
    navigate(`/blocks/${index}`);
  };

  return (
    <div className="space-y-8">

      {/* TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiPackage size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">Blockchain Blocks</h2>
      </div>

      {/* LOADING */}
      {loading && (
        <p className="text-gray-300 animate-pulse text-center">Loading blocks…</p>
      )}

      {/* ERROR */}
      {error && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded border border-red-700 text-center">
          {error}
        </p>
      )}

      {/* BLOCK TABLE */}
      {!loading && !error && (
        <div className="glass rounded-2xl shadow-xl overflow-hidden">

          {blocks.length === 0 ? (
            <p className="text-gray-400 p-6 text-center">No blocks found.</p>
          ) : (
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gradient-to-r from-primary to-secondary text-black">
                  <th className="py-3 px-4 text-left text-sm font-semibold">Index</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Hash</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Prev Hash</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Timestamp</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Tx Count</th>
                </tr>
              </thead>

              <tbody>
                {blocks.map((blk) => {
                  const date = blk.timestamp
                    ? new Date(blk.timestamp * 1000).toLocaleString()
                    : "";

                  return (
                    <tr
                      key={blk.index}
                      className="border-t border-gray-700 hover:bg-white/5 cursor-pointer transition-all"
                      onClick={() => handleRowClick(blk.index)}
                    >
                      <td className="py-3 px-4 text-gray-300">{blk.index}</td>

                      <td className="py-3 px-4 text-gray-300 font-mono text-xs break-all">
                        {blk.hash}
                      </td>

                      <td className="py-3 px-4 text-gray-300 font-mono text-xs break-all">
                        {blk.prev_hash}
                      </td>

                      <td className="py-3 px-4 text-gray-400">{date}</td>

                      <td className="py-3 px-4 text-lg font-bold text-secondary">
                        {blk.tx_count}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
