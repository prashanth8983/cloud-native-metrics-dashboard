import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
  BarChart, 
  PieChart, 
  AlertCircle, 
  Search, 
  Settings, 
  ChevronLeft, 
  ChevronRight,
  Home,
  Code,
  Server,
  Database
} from 'lucide-react';
//import { useTheme } from '../../context/ThemeContext';



interface SidebarProps {
  collapsed: boolean;
  toggleCollapsed: () => void;
}

const Sidebar: React.FC<SidebarProps> = ({ collapsed, toggleCollapsed }) => {
  //const { theme } = useTheme();
  const location = useLocation();
  
  const navItems = [
    { path: '/', label: 'Dashboard', icon: <Home size={20} /> },
    { path: '/metrics', label: 'Metrics Explorer', icon: <BarChart size={20} /> },
    { path: '/queries', label: 'Query Builder', icon: <Search size={20} /> },
    { path: '/alerts', label: 'Alerts', icon: <AlertCircle size={20} /> },
    { path: '/services', label: 'Services', icon: <Server size={20} /> },
    { path: '/databases', label: 'Databases', icon: <Database size={20} /> },
    { path: '/traces', label: 'Traces', icon: <PieChart size={20} /> },
    { path: '/logs', label: 'Logs', icon: <Code size={20} /> },
    { path: '/settings', label: 'Settings', icon: <Settings size={20} /> },
  ];

  return (
    <div
      className={`flex flex-col h-screen bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 transition-all duration-300 ${
        collapsed ? 'w-16' : 'w-64'
      }`}
    >
      <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-800">
        {!collapsed && (
          <h2 className="text-xl font-bold text-gray-800 dark:text-white">
            Metrics<span className="text-blue-600">Dashboard</span>
          </h2>
        )}
        <button
          onClick={toggleCollapsed}
          className="p-1 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400"
        >
          {collapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
        </button>
      </div>
      
      <nav className="flex-1 py-4 overflow-y-auto">
        <ul className="space-y-1">
          {navItems.map((item) => (
            <li key={item.path}>
              <Link
                to={item.path}
                className={`flex items-center px-4 py-3 text-gray-700 dark:text-gray-200 rounded-md hover:bg-gray-100 dark:hover:bg-gray-800 ${
                  location.pathname === item.path ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400' : ''
                }`}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {!collapsed && <span className="ml-3">{item.label}</span>}
              </Link>
            </li>
          ))}
        </ul>
      </nav>
      
      <div className="p-4 border-t border-gray-200 dark:border-gray-800">
        <div className="flex items-center">
          {!collapsed && (
            <>
              <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-bold">
                U
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-700 dark:text-gray-200">User Name</p>
                <p className="text-xs text-gray-500 dark:text-gray-400">Administrator</p>
              </div>
            </>
          )}
          {collapsed && (
            <div className="w-8 h-8 rounded-full bg-blue-500 flex items-center justify-center text-white font-bold">
              U
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Sidebar;