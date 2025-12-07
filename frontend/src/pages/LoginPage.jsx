import React, { useState } from "react";
import { useNavigate, Link } from "react-router-dom";
import { requestOtp, verifyOtp } from "../api/auth.js";
import { useAuth } from "../context/AuthContext.jsx";

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [otpSent, setOtpSent] = useState(false);
  const [serverOtp, setServerOtp] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const { setAuthData } = useAuth();
  const navigate = useNavigate();

  const handleRequestOtp = async (e) => {
    e.preventDefault();
    setError("");
    if (!email) {
      setError("Please enter your email.");
      return;
    }
    setLoading(true);
    try {
      const resp = await requestOtp(email);
      setOtpSent(true);
      setServerOtp(resp.otp);
    } catch (err) {
      const message = err?.message || err?.error || "Failed to request OTP";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyOtp = async (e) => {
    e.preventDefault();
    setError("");
    if (!otp) {
      setError("Please enter the OTP.");
      return;
    }
    setLoading(true);
    try {
      const resp = await verifyOtp(email, otp);
      if (resp.success) {
        const savedWalletAddress = localStorage.getItem("walletAddress");
        const savedPrivateKey = localStorage.getItem("privateKey");

        setAuthData({
          email,
          walletAddress: savedWalletAddress || null,
          privateKey: savedPrivateKey || null,
        });

        navigate("/dashboard");
      } else {
        setError(resp.message || "Invalid or expired OTP");
      }
    } catch (err) {
      const message = err?.message || err?.error || "Failed to verify OTP";
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex justify-center items-center bg-gradient-to-br from-gray-900 via-black to-gray-800 px-4">

      <div className="glass w-full max-w-md p-8 shadow-xl rounded-2xl">

        <h2 className="text-3xl font-bold text-primary text-center mb-6">
          Login
        </h2>

        {!otpSent ? (
          // STEP 1 — Request OTP
          <form onSubmit={handleRequestOtp} className="space-y-5">

            <div>
              <label htmlFor="email" className="text-sm text-gray-300">
                Email Address
              </label>
              <input
                id="email"
                type="email"
                className="mt-1 w-full p-3 rounded bg-gray-800 border border-gray-700 focus:border-primary focus:ring-primary text-gray-100"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>

            {error && <p className="text-red-400 text-sm">{error}</p>}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded shadow-md disabled:opacity-50"
            >
              {loading ? "Requesting…" : "Request OTP"}
            </button>

          </form>
        ) : (
          // STEP 2 — Verify OTP
          <form onSubmit={handleVerifyOtp} className="space-y-5">

            <div className="p-3 rounded bg-gray-800 border border-gray-700 text-sm">
              <p className="text-gray-300">
                Demo OTP:
                <span className="font-mono font-bold text-primary ml-2">
                  {serverOtp}
                </span>
              </p>
            </div>

            <div>
              <label htmlFor="otp" className="text-sm text-gray-300">
                Enter OTP
              </label>
              <input
                id="otp"
                type="text"
                className="mt-1 w-full p-3 rounded bg-gray-800 border border-gray-700 focus:border-primary focus:ring-primary text-gray-100"
                value={otp}
                onChange={(e) => setOtp(e.target.value)}
                required
              />
            </div>

            {error && <p className="text-red-400 text-sm">{error}</p>}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 bg-primary hover:bg-secondary text-black font-semibold rounded shadow-md disabled:opacity-50"
            >
              {loading ? "Verifying…" : "Verify OTP"}
            </button>
          </form>
        )}

        <p className="mt-6 text-sm text-gray-400 text-center">
          Don't have an account?{" "}
          <Link to="/register" className="text-primary hover:underline">
            Register here
          </Link>
        </p>

      </div>
    </div>
  );
}
