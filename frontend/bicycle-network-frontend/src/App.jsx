import { Routes, Route, Navigate } from 'react-router-dom'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import Register from './pages/Register'
import ForgotPassword from './pages/ForgotPassword'
import ResetPassword from './pages/ResetPassword'
import Dashboard from './pages/Dashboard'
import Bicycles from './pages/Bicycles'
import BikeDetail from './pages/BikeDetail'
import MapPage from './pages/MapPage'
import Events from './pages/Events'
import EventDetail from './pages/EventDetail'
import Profile from './pages/Profile'
import GoogleCallback from './pages/GoogleCallback'
import { useAuth } from './contexts/AuthContext'
import './App.css'

function App() {
  const { loading } = useAuth()

  if (loading) {
    return (
      <div className="loading-screen">
        <div className="spinner" />
      </div>
    )
  }

  return (
    <Routes>
      {/* Public routes */}
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />
      <Route path="/auth/callback" element={<GoogleCallback />} />

      {/* Protected routes */}
      <Route
        element={
          <ProtectedRoute>
            <Layout />
          </ProtectedRoute>
        }
      >
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/bicycles" element={<Bicycles />} />
        <Route path="/bicycles/:id" element={<BikeDetail />} />
        <Route path="/map" element={<MapPage />} />
        <Route path="/events" element={<Events />} />
        <Route path="/events/:id" element={<EventDetail />} />
        <Route path="/profile" element={<Profile />} />
      </Route>

      {/* Redirect */}
      <Route path="*" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

export default App
