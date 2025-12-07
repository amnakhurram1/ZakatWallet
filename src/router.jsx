import React from 'react';
import {
  BrowserRouter,
  Routes,
  Route,
} from 'react-router-dom';
import Layout from './components/Layout.jsx';
import ProtectedRoute from './components/ProtectedRoute.jsx';
import LoginPage from './pages/LoginPage.jsx';
import RegisterPage from './pages/RegisterPage.jsx';
import DashboardPage from './pages/DashboardPage.jsx';
import SendPage from './pages/SendPage.jsx';
import HistoryPage from './pages/HistoryPage.jsx';
import BlocksPage from './pages/BlocksPage.jsx';
import BlockDetailPage from './pages/BlockDetailPage.jsx';
import ZakatPage from './pages/ZakatPage.jsx';
import LogsPage from './pages/LogsPage.jsx';
import AdminFundPage from './pages/AdminFundPage.jsx';

// Defines the routing structure for the application.  Public pages
// (login/register) are available without authentication.  The
// ProtectedRoute component can later enforce auth for private
// routes; currently it simply renders its children.
export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}> 
          {/* Public routes */}
          <Route index element={<LoginPage />} />
          <Route path="login" element={<LoginPage />} />
          <Route path="register" element={<RegisterPage />} />

          {/* Protected routes */}
          <Route
            path="dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="send"
            element={
              <ProtectedRoute>
                <SendPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="history"
            element={
              <ProtectedRoute>
                <HistoryPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="blocks"
            element={
              <ProtectedRoute>
                <BlocksPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="blocks/:index"
            element={
              <ProtectedRoute>
                <BlockDetailPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="zakat"
            element={
              <ProtectedRoute>
                <ZakatPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="logs"
            element={
              <ProtectedRoute>
                <LogsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="admin/fund"
            element={
              <ProtectedRoute>
                <AdminFundPage />
              </ProtectedRoute>
            }
          />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
