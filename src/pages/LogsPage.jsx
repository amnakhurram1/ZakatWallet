import React, { useEffect, useState } from "react";
import { getSystemLogs } from "../api/logs.js";
import {
  FiTerminal,
  FiInfo,
  FiAlertTriangle,
  FiAlertCircle,
} from "react-icons/fi";

export default function LogsPage() {
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    getSystemLogs(100)
      .then((data) => {
        setLogs(Array.isArray(data?.logs) ? data.logs : []);
      })
      .catch((err) => {
        setError(err?.message || err?.error || String(err));
      })
      .finally(() => setLoading(false));
  }, []);

  const renderLevelIcon = (level) => {
    const l = level?.toLowerCase();
    if (l === "warn" || l === "warning")
      return <FiAlertTriangle className="text-yellow-400" size={20} />;
    if (l === "error")
      return <FiAlertCircle className="text-red-400" size={20} />;
    return <FiInfo className="text-blue-400" size={20} />;
  };

  const getLevelColor = (level) => {
    const l = level?.toLowerCase();
    if (l === "warn" || l === "warning") return "text-yellow-400";
    if (l === "error") return "text-red-400";
    return "text-blue-400";
  };

  return (
    <div className="space-y-8">

      {/* PAGE TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiTerminal size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">System Logs</h2>
      </div>

      {/* LOADING */}
      {loading && (
        <p className="text-gray-300 animate-pulse text-center">Loading logs…</p>
      )}

      {/* ERROR */}
      {error && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded border border-red-700 text-center">
          {error}
        </p>
      )}

      {/* LOGS TABLE */}
      {!loading && !error && (
        <div className="glass rounded-2xl shadow-xl overflow-hidden">
          {logs.length === 0 ? (
            <p className="text-gray-400 p-6 text-center">No logs available.</p>
          ) : (
            <table className="w-full border-collapse">
              <thead>
                <tr className="bg-gradient-to-r from-primary to-secondary text-black">
                  <th className="py-3 px-4 text-left text-sm font-semibold">Timestamp</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Level</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Type</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">Message</th>
                  <th className="py-3 px-4 text-left text-sm font-semibold">IP</th>
                </tr>
              </thead>

              <tbody>
                {logs.map((log, idx) => (
                  <tr
                    key={idx}
                    className="border-t border-gray-700 hover:bg-white/5 transition-all"
                  >
                    {/* TIMESTAMP */}
                    <td className="py-3 px-4 text-gray-300">
                      {log.timestamp
                        ? new Date(log.timestamp).toLocaleString()
                        : ""}
                    </td>

                    {/* LEVEL WITH ICON */}
                    <td className="py-3 px-4 flex items-center gap-2">
                      {renderLevelIcon(log.level)}
                      <span className={`${getLevelColor(log.level)} font-semibold`}>
                        {log.level}
                      </span>
                    </td>

                    {/* TYPE */}
                    <td className="py-3 px-4 text-gray-300">{log.type}</td>

                    {/* MESSAGE */}
                    <td className="py-3 px-4 text-gray-400 break-words max-w-lg">
                      {log.message}
                    </td>

                    {/* IP */}
                    <td className="py-3 px-4 text-gray-300">{log.ip}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}
    </div>
  );
}
