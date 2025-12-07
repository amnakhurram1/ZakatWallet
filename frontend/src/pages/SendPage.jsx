import React, { useState } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import { sendTransaction } from "../api/transactions.js";
import { FiSend } from "react-icons/fi";

export default function SendPage() {
  const { walletAddress, privateKey } = useAuth();
  const [toAddress, setToAddress] = useState("");
  const [amount, setAmount] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(null);
    setMessage(null);

    // Validation
    if (!toAddress.trim()) {
      setError("Recipient address is required");
      return;
    }
    const amt = Number(amount);
    if (isNaN(amt) || amt <= 0) {
      setError("Amount must be a positive number");
      return;
    }
    if (!walletAddress || !privateKey) {
      setError("Missing wallet or private key");
      return;
    }

    setLoading(true);
    try {
      const res = await sendTransaction({
        from: walletAddress,
        to: toAddress.trim(),
        amount: amt,
        privKey: privateKey,
      });

      setMessage(res?.status || "Transaction submitted successfully!");
      setToAddress("");
      setAmount("");
    } catch (err) {
      setError(err?.message || err?.error || String(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto glass p-8 rounded-2xl shadow-xl space-y-6">

      {/* PAGE TITLE */}
      <div className="flex items-center gap-3 mb-6">
        <FiSend size={32} className="text-primary" />
        <h2 className="text-3xl font-bold text-primary">Send Funds</h2>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">

        {/* TO ADDRESS */}
        <div>
          <label className="block text-sm text-gray-300 mb-1">
            Recipient Wallet Address
          </label>
          <input
            type="text"
            className="w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary focus:ring-primary"
            value={toAddress}
            onChange={(e) => setToAddress(e.target.value)}
            placeholder="Enter wallet address"
          />
        </div>

        {/* AMOUNT */}
        <div>
          <label className="block text-sm text-gray-300 mb-1">
            Amount
          </label>
          <input
            type="number"
            className="w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="Units to send"
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
          {loading ? "Sending…" : "Send Transaction"}
        </button>
      </form>

      {/* SUCCESS MESSAGE */}
      {message && (
        <p className="text-green-400 bg-green-900/20 p-3 rounded mt-2 text-sm border border-green-700">
          {message}
        </p>
      )}

      {/* ERROR MESSAGE */}
      {error && (
        <p className="text-red-400 bg-red-900/20 p-3 rounded mt-2 text-sm border border-red-700">
          {error}
        </p>
      )}
    </div>
  );
}
