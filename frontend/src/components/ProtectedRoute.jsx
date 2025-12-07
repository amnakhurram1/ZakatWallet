import React from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext.jsx";

export default function ProtectedRoute({ children }) {
  const { email } = useAuth();
  const location = useLocation();

  // Consider the user "logged in" if we have an email.
  if (!email) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}
