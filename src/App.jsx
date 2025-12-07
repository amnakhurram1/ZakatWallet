import React from 'react';
import Router from './router.jsx';
import { AuthProvider } from './context/AuthContext.jsx';

// Top‑level component that renders the router.  Additional
// providers (e.g. for state management) can be added here later.
export default function App() {
  // Wrap the router with the authentication provider so that all
  // components have access to authentication state.  Additional
  // providers (e.g. for global state management) can be nested here.
  return (
    <AuthProvider>
      <Router />
    </AuthProvider>
  );
}
