import React, { useEffect, useState } from "react";
import { useAuth } from "../context/AuthContext.jsx";
import { getBalance, getWalletReport } from "../api/wallet.js";

// NO MORE react-icons IMPORTS

export default function DashboardPage() {
  const { walletAddress } = useAuth();
  const [balance, setBalance] = useState(null);
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let mounted = true;

    const fetchData = async () => {
      if (!walletAddress) {
        setError("No wallet address available.");
        setLoading(false);
        return;
      }

      setLoading(true);
      setError("");

      try {
        const balResp = await getBalance(walletAddress);
        const repResp = await getWalletReport(walletAddress);

        if (mounted) {
          setBalance(balResp?.balance);
          setReport(repResp);
        }
      } catch (err) {
        const msg = err?.error || err?.message || "Failed to fetch data";
        if (mounted) setError(msg);
      } finally {
        if (mounted) setLoading(false);
      }
    };

    fetchData();
    return () => {
      mounted = false;
    };
  }, [walletAddress]);

  if (loading) {
    return (
      <div className="text-center text-gray-300 animate-pulse text-lg">
        Loading dashboard…
      </div>
    );
  }

  if (error) {
    return <div className="text-red-400 text-center text-lg">{error}</div>;
  }

  if (!report) {
    return <div className="text-gray-400 text-center">No report available.</div>;
  }

  // REPLACED ICONS WITH EMOJIS
  const dashboardCards = [
    {
      title: "Current Balance",
      value: balance ?? report.balance,
      icon: "💰",
      color: "from-blue-500 to-cyan-400",
    },
    {
      title: "Total Sent",
      value: report.total_sent,
      icon: "📤",
      color: "from-red-500 to-orange-400",
    },
    {
      title: "Total Received",
      value: report.total_received,
      icon: "📥",
      color: "from-green-500 to-emerald-400",
    },
    {
      title: "Total Zakat",
      value: report.total_zakat,
      icon: "❤️",
      color: "from-purple-500 to-pink-400",
    },
  ];

  return (
    <div className="space-y-8">
      <h2 className="text-4xl font-bold text-primary drop-shadow mb-8">
        Dashboard Overview
      </h2>

      {/* CARDS GRID */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {dashboardCards.map((card, idx) => (
          <div
            key={idx}
            className={`glass p-6 rounded-2xl shadow-lg flex flex-col gap-4 transform hover:scale-105 transition-all duration-300`}
          >
            {/* ICON */}
            <div
              className={`w-14 h-14 rounded-xl bg-gradient-to-br ${card.color} flex justify-center items-center text-white shadow-md text-3xl`}
            >
              {card.icon}
            </div>

            {/* TEXT */}
            <h3 className="text-gray-300 text-sm uppercase tracking-wide">
              {card.title}
            </h3>

            <p className="text-3xl font-extrabold text-white">{card.value}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
