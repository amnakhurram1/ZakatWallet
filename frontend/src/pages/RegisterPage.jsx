import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { registerUser } from "../api/auth.js";
import { useAuth } from "../context/AuthContext.jsx";

export default function RegisterPage() {
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [cnic, setCnic] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState(null);

  const { setAuthData } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const resp = await registerUser({ fullName, email, cnic });
      setResult(resp);

      // Save email + wallet details in global auth state
      setAuthData({
        email: resp.email,
        walletAddress: resp.wallet_address,
        privateKey: resp.private_key,
      });
    } catch (err) {
      const message = err?.message || err?.error || "Failed to register";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex justify-center items-center bg-gradient-to-br from-gray-900 via-black to-gray-800 px-4">

      <div className="glass w-full max-w-md p-8 shadow-xl rounded-2xl">

        <h2 className="text-3xl font-bold text-primary text-center mb-6">
          Create Account
        </h2>

        {result ? (
          // ============================
          // SUCCESS SCREEN
          // ============================
          <div className="space-y-5">

            <p className="text-green-400 text-lg font-semibold">
              Registration successful!
            </p>

            <div className="p-4 rounded-xl bg-gray-800 border border-gray-700 shadow">
              <p className="text-gray-300 mb-2">
                <span className="font-semibold text-primary">Wallet Address:</span>
                <br />
                <span className="font-mono break-all">
                  {result.wallet_address}
                </span>
              </p>

              <p className="text-gray-300 mt-4">
                <span className="font-semibold text-primary">Private Key:</span>
                <br />
                <span className="font-mono break-all text-red-400">
                  {result.private_key}
                </span>
              </p>
            </div>

            <p className="text-gray-400 text-sm">
              ⚠️ Save your private key securely.  
              It will NOT be shown again.
            </p>

            <button
              onClick={() => navigate("/dashboard")}
              className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded shadow-md"
            >
              Go to Dashboard
            </button>

          </div>
        ) : (
          // ============================
          // REGISTRATION FORM
          // ============================
          <form onSubmit={handleSubmit} className="space-y-5">

            <div>
              <label htmlFor="fullName" className="text-sm text-gray-300">
                Full Name
              </label>
              <input
                id="fullName"
                type="text"
                className="mt-1 w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary focus:ring-primary"
                value={fullName}
                onChange={(e) => setFullName(e.target.value)}
                required
              />
            </div>

            <div>
              <label htmlFor="email" className="text-sm text-gray-300">
                Email
              </label>
              <input
                id="email"
                type="email"
                className="mt-1 w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary focus:ring-primary"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            <div>
              <label htmlFor="cnic" className="text-sm text-gray-300">
                CNIC
              </label>
              <input
                id="cnic"
                type="text"
                className="mt-1 w-full p-3 rounded bg-gray-800 border border-gray-700 text-gray-100 focus:border-primary focus:ring-primary"
                value={cnic}
                onChange={(e) => setCnic(e.target.value)}
                required
              />
            </div>

            {error && <p className="text-red-400 text-sm">{error}</p>}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded shadow-md disabled:opacity-50"
            >
              {loading ? "Registering…" : "Register"}
            </button>
          </form>
        )}

        {!result && (
          <p className="mt-6 text-sm text-gray-400 text-center">
            Already have an account?{" "}
            <Link to="/login" className="text-primary hover:underline">
              Login here
            </Link>
          </p>
        )}

      </div>
    </div>
  );
}
