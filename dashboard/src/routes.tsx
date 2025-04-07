import React, { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Loader } from './components/ui';

// Lazy-loaded components
const Dashboard = lazy(() => import('./pages/Dashboard'));
const MetricsExplorer = lazy(() => import('./pages/MetricsExplorer'));

// Placeholder components for future implementation
const QueryBuilder = () => <h1>Query Builder (Coming Soon)</h1>;
const Alerts = () => <h1>Alerts (Coming Soon)</h1>;
const Settings = () => <h1>Settings (Coming Soon)</h1>;

const Loading = () => (
  <div className="flex items-center justify-center h-screen bg-gray-50 dark:bg-gray-900">
    <Loader size="lg" text="Loading..." />
  </div>
);

const AppRoutes: React.FC = () => {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/metrics" element={<MetricsExplorer />} />
        <Route path="/queries" element={<QueryBuilder />} />
        <Route path="/alerts" element={<Alerts />} />
        <Route path="/settings" element={<Settings />} />
        
        {/* Catch-all redirect to dashboard */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  );
};

export default AppRoutes;