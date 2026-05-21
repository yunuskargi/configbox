import { Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Devices from './pages/Devices';
import Backups from './pages/Backups';
import Locations from './pages/Locations';
import Users from './pages/Users';
import SmtpSettings from './pages/SmtpSettings';
import Settings from './pages/Settings';
import AuditLog from './pages/AuditLog';
import RemoteBackup from './pages/RemoteBackup';

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/devices" element={<Devices />} />
        <Route path="/backups" element={<Backups />} />
        <Route path="/locations" element={<Locations />} />
        <Route path="/users" element={<Users />} />
        <Route path="/smtp" element={<SmtpSettings />} />
        <Route path="/remote-backup" element={<RemoteBackup />} />
        <Route path="/audit" element={<AuditLog />} />
        <Route path="/settings" element={<Settings />} />
      </Route>
    </Routes>
  );
}
