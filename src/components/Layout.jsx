import React from "react";
import { Link, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext.jsx";

export default function Layout() {
  const { email, clearAuthData } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    clearAuthData();
    navigate("/login");
  };

  return (
    <div className="min-h-screen flex bg-gradient-to-br from-gray-900 via-black to-gray-800 text-gray-100">

      {/* SIDEBAR */}
      <aside className="w-72 glass p-6 flex flex-col gap-8 shadow-xl sticky top-0 h-screen">

        {/* Branding */}
        <h1 className="text-3xl font-bold tracking-wide text-primary">
          Crypto Wallet
        </h1>

        {/* NAVIGATION */}
        <nav className="flex flex-col gap-4 text-lg font-medium">
          {email ? (
            <>
              <Link className="hover:text-primary" to="/dashboard">
                Dashboard
              </Link>
              <Link className="hover:text-primary" to="/send">
                Send
              </Link>
              <Link className="hover:text-primary" to="/history">
                History
              </Link>
              <Link className="hover:text-primary" to="/blocks">
                Blocks
              </Link>
              <Link className="hover:text-primary" to="/zakat">
                Zakat
              </Link>
              <Link className="hover:text-primary" to="/logs">
                Logs
              </Link>
              <Link className="hover:text-primary" to="/admin/fund">
                Admin Fund
              </Link>

              {/* LOGOUT BUTTON */}
              <button
                onClick={handleLogout}
                className="text-left hover:text-red-400 mt-4"
              >
                Logout
              </button>
            </>
          ) : (
            <>
              <Link className="hover:text-primary" to="/login">
                Login
              </Link>
              <Link className="hover:text-primary" to="/register">
                Register
              </Link>
            </>
          )}
        </nav>
      </aside>

      {/* MAIN CONTENT */}
      <main className="flex-1 p-10">
        <div className="glass p-10 shadow-xl rounded-2xl">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
