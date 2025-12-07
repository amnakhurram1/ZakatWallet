import React, { createContext, useContext, useEffect, useState } from "react";

// Defines an authentication context for the wallet application.  It
// stores the user's email address, wallet address and private key
// (where available), initialising values from localStorage on load
// and synchronising changes back to localStorage.  Components can
// consume this context via the useAuth() hook.

const AuthContext = createContext({
  email: null,
  walletAddress: null,
  privateKey: null,
  setAuthData: () => {},
  clearAuthData: () => {},
});

export function AuthProvider({ children }) {
  const [authData, setAuthDataState] = useState(() => {
    // Initialise state from localStorage if present.  Keys are
    // case-sensitive to avoid collisions with other storage.
    const email = localStorage.getItem("email");
    const walletAddress = localStorage.getItem("walletAddress");
    const privateKey = localStorage.getItem("privateKey");
    return {
      email: email || null,
      walletAddress: walletAddress || null,
      privateKey: privateKey || null,
    };
  });

  // Persist authData to localStorage whenever it changes.
  useEffect(() => {
    if (authData.email) {
      localStorage.setItem("email", authData.email);
    } else {
      localStorage.removeItem("email");
    }
    if (authData.walletAddress) {
      localStorage.setItem("walletAddress", authData.walletAddress);
    } else {
      localStorage.removeItem("walletAddress");
    }
    if (authData.privateKey) {
      localStorage.setItem("privateKey", authData.privateKey);
    } else {
      localStorage.removeItem("privateKey");
    }
  }, [authData]);

  // Update context state with new auth information.  Fields that are
  // undefined keep their previous value; fields that are explicitly
  // null are cleared.
  const setAuthData = ({ email, walletAddress, privateKey }) => {
    setAuthDataState((prev) => ({
      email: email !== undefined ? email : prev.email,
      walletAddress:
        walletAddress !== undefined ? walletAddress : prev.walletAddress,
      privateKey: privateKey !== undefined ? privateKey : prev.privateKey,
    }));
  };

  // Clear all stored authentication information.
  const clearAuthData = () => {
    setAuthDataState({ email: null, walletAddress: null, privateKey: null });
  };

  return (
    <AuthContext.Provider value={{ ...authData, setAuthData, clearAuthData }}>
      {children}
    </AuthContext.Provider>
  );
}

// Convenience hook for consuming the AuthContext.  Always returns
// defined properties to avoid undefined errors in consumers.
export function useAuth() {
  return useContext(AuthContext);
}
