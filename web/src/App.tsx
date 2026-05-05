import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Users from './pages/Users';
import Groups from './pages/Groups';
import Audit from './pages/Audit';
import Inactive from './pages/Inactive';
import Reports from './pages/Reports';
import Settings from './pages/Settings';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="/users" element={<Users />} />
        <Route path="/inactive" element={<Inactive />} />
        <Route path="/groups" element={<Groups />} />
        <Route path="/reports" element={<Reports />} />
        <Route path="/audit" element={<Audit />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
