import React, { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import { getWalletReport } from "../api/wallet.js";
import { runZakat } from "../api/zakat.js";
import { FiActivity, FiHeart, FiDatabase } from "react-icons/fi";

export default function ZakatPage() {
  const { walletAddress } = useAuth();

  const [report, setReport] = useState(null);
  const [loadingReport, setLoadingReport] = useState(true);
  const [errorReport, setErrorReport] = useState(null);

  const [running, setRunning] = useState(false);
  const [zakatResult, setZakatResult] = useState(null);
  const [runError, setRunError] = useState(null);

  const fetchReport = () => {
    if (!walletAddress) return;
    setLoadingReport(true);
    setErrorReport(null);

    getWalletReport(walletAddress)
      .then((data) => setReport(data))
      .catch((err) =>
        setErrorReport(err?.message || err?.error || String(err))
      )
      .finally(() => setLoadingReport(false));
  };

  useEffect(() => {
    fetchReport();
  }, [walletAddress]);

  const handleRunZakat = async () => {
    setRunning(true);
    setRunError(null);
    setZakatResult(null);

    try {
      const result = await runZakat();
      setZakatResult(result);
      fetchReport();
    } catch (err) {
      setRunError(err?.message || err?.error || String(err));
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto space-y-8">

      {/* TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiHeart size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">Zakat Management</h2>
      </div>

      {/* WALLET SUMMARY */}
      {loadingReport && (
        <p className="text-gray-300 animate-pulse">Loading your report…</p>
      )}

      {errorReport && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded border border-red-700">
          {errorReport}
        </p>
      )}

      {!loadingReport && !errorReport && report && (
        <div className="glass p-6 rounded-2xl shadow-lg space-y-3">

          <div className="flex items-center gap-3">
            <FiActivity className="text-secondary" size={28} />
            <h3 className="text-xl font-semibold text-gray-200">
              Your Wallet Summary
            </h3>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
              <p className="text-gray-400 text-sm">Current Balance</p>
              <p className="text-2xl font-bold text-primary">
                {report.balance}
              </p>
            </div>

            <div className="p-4 bg-gray-800 rounded-lg border border-gray-700">
              <p className="text-gray-400 text-sm">Total Zakat Paid</p>
              <p className="text-2xl font-bold text-pink-400">
                {report.total_zakat}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* RUN ZAKAT BUTTON */}
      <button
        onClick={handleRunZakat}
        disabled={running}
        className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded-lg shadow-lg disabled:opacity-50 transform hover:scale-[1.02] transition-all duration-300"
      >
        {running ? "Running Zakat…" : "Run Global Zakat Now"}
      </button>

      {/* ERROR AFTER RUN */}
      {runError && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded border border-red-700">
          {runError}
        </p>
      )}

      {/* ZAKAT RESULT */}
      {zakatResult && (
        <div className="glass p-6 rounded-2xl shadow-lg space-y-4">

          {/* HEADER */}
          <div className="flex items-center gap-3">
            <FiDatabase className="text-secondary" size={28} />
            <h3 className="text-xl font-semibold text-gray-200">
              Zakat Run Result
            </h3>
          </div>

          {/* METRICS */}
          <div className="grid md:grid-cols-3 gap-4">
            <div className="p-4 bg-gray-800 border border-gray-700 rounded-lg">
              <p className="text-gray-400 text-sm">Total Wallets</p>
              <p className="text-xl font-bold text-primary">
                {zakatResult.total_wallets}
              </p>
            </div>

            <div className="p-4 bg-gray-800 border border-gray-700 rounded-lg">
              <p className="text-gray-400 text-sm">Processed</p>
              <p className="text-xl font-bold text-green-400">
                {zakatResult.processed}
              </p>
            </div>

            <div className="p-4 bg-gray-800 border border-gray-700 rounded-lg">
              <p className="text-gray-400 text-sm">Total Zakat</p>
              <p className="text-xl font-bold text-pink-400">
                {zakatResult.total_zakat}
              </p>
            </div>
          </div>

          {/* BLOCK HASHES */}
          <div className="mt-4">
            <p className="font-semibold text-gray-300 mb-2">Block Hashes:</p>
            <ul className="list-disc list-inside text-gray-400 text-sm space-y-1">
              {Array.isArray(zakatResult.block_hashes) &&
                zakatResult.block_hashes.map((hash) => (
                  <li
                    key={hash}
                    className="font-mono break-all bg-gray-900/40 p-2 rounded border border-gray-700"
                  >
                    {hash}
                  </li>
                ))}
            </ul>
          </div>

        </div>
      )}
    </div>
  );
}
