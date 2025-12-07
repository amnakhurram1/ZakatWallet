import React, { useState } from "react";
import { fundWallet } from "../api/admin.js";
import { FiCreditCard, FiCheckCircle, FiAlertCircle } from "react-icons/fi";

export default function AdminFundPage() {
  const [address, setAddress] = useState("");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setResult(null);

    const amtNum = Number(amount);

    if (!address.trim()) {
      setError("Address is required");
      return;
    }

    if (isNaN(amtNum) || amtNum <= 0) {
      setError("Amount must be a positive number");
      return;
    }

    setLoading(true);
    try {
      const res = await fundWallet({
        address: address.trim(),
        amount: amtNum,
      });
      setResult(res);
      setAddress("");
      setAmount("");
    } catch (err) {
      setError(err?.message || err?.error || String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-xl mx-auto space-y-8">

      {/* TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiCreditCard size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">Admin Fund Wallet</h2>
      </div>

      {/* FORM */}
      <div className="glass p-8 rounded-2xl shadow-xl">
        <form onSubmit={handleSubmit} className="space-y-6">

          {/* ADDRESS FIELD */}
          <div>
            <label className="block text-sm text-gray-300 mb-1">
              Recipient Wallet Address
            </label>
            <input
              type="text"
              className="w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary focus:ring-primary"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="Enter wallet address"
            />
          </div>

          {/* AMOUNT FIELD */}
          <div>
            <label className="block text-sm text-gray-300 mb-1">Amount</label>
            <input
              type="number"
              className="w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              placeholder="Amount to fund"
              min="0"
              step="1"
            />
          </div>

          {/* SUBMIT BUTTON */}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded shadow-lg disabled:opacity-50 transform hover:scale-[1.02] transition-all duration-300"
          >
            {loading ? "Funding…" : "Fund Wallet"}
          </button>
        </form>

        {/* ERROR MESSAGE */}
        {error && (
          <div className="mt-6 flex gap-3 items-center p-4 bg-red-900/20 border border-red-700 text-red-300 rounded-lg">
            <FiAlertCircle size={22} className="flex-shrink-0" />
            <p className="text-sm">{error}</p>
          </div>
        )}

        {/* SUCCESS MESSAGE */}
        {result && (
          <div className="mt-6 glass p-5 rounded-xl border border-green-700 text-green-300 shadow-lg">
            <div className="flex items-center gap-3 mb-2">
              <FiCheckCircle size={24} className="text-green-400" />
              <p className="font-semibold text-lg">Wallet funded successfully!</p>
            </div>
            <p className="text-sm text-gray-200">
              <span className="font-semibold text-green-400">Block Hash:</span>
              <br />
              <span className="font-mono break-all">{result.block_hash}</span>
            </p>
          </div>
        )}
      </div>

    </div>
  );
}
