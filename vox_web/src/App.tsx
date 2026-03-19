import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './hooks/useAuth';
import { ProtectedRoute } from './components/ProtectedRoute';

import { HomePage } from './pages/HomePage';
import { HostDashboardPage } from './pages/HostDashboardPage';
import { HostBroadcastPage } from './pages/HostBroadcastPage';
import { ListenerPage } from './pages/ListenerPage';
import { ProfilePage } from './pages/ProfilePage';

import './styles/globals.css';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Public */}
          <Route path="/" element={<HomePage />} />
          <Route path="/listen" element={<ListenerPage />} />

          {/* Protected — host only */}
          <Route
            path="/host"
            element={
              <ProtectedRoute>
                <HostDashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/host/:hubId"
            element={
              <ProtectedRoute>
                <HostBroadcastPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/profile"
            element={
              <ProtectedRoute>
                <ProfilePage />
              </ProtectedRoute>
            }
          />

          {/* Fallback */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
