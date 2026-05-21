import { Outlet, Navigate, useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) return <div className="flex items-center justify-center min-h-screen text-gray-500 dark:bg-gray-950 dark:text-gray-400">Loading...</div>;
  if (!user) return <Navigate to="/login" />;

  // Force password change for default admin
  if (user.must_change_password && location.pathname !== '/settings') {
    return <Navigate to="/settings" replace />;
  }

  return (
    <div className="flex min-h-screen bg-gray-50 dark:bg-gray-950">
      <Sidebar />
      <main className="flex-1 p-6 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
