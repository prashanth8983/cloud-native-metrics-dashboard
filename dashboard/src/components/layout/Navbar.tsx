import React, { useState } from 'react';
import { Bell, Sun, Moon, Search, HelpCircle, Menu } from 'lucide-react';
import { useTheme } from '../../context/ThemeContext';
import { useTimeRangeContext } from '../../context/TimeRangeContext';
import { TIME_RANGE_PRESETS } from '../../utils/constant';
import { TimeRangePreset } from '../../types/common';



interface NavbarProps {
  toggleSidebar: () => void;
}

const Navbar: React.FC<NavbarProps> = ({ toggleSidebar }) => {
  const { theme, toggleTheme } = useTheme();
  const { preset, setTimeRangeFromPreset } = useTimeRangeContext();
  const [searchOpen, setSearchOpen] = useState(false);

  return (
    <header className="bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800">
      <div className="px-4 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={toggleSidebar}
            className="p-1 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400 md:hidden"
          >
            <Menu size={20} />
          </button>
          
          {searchOpen ? (
            <div className="relative">
              <input
                type="text"
                placeholder="Search metrics..."
                className="w-64 px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-md dark:bg-gray-800 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                autoFocus
                onBlur={() => setSearchOpen(false)}
              />
              <button
                className="absolute right-2 top-1/2 transform -translate-y-1/2 text-gray-500"
                onClick={() => setSearchOpen(false)}
              >
                <Search size={16} />
              </button>
            </div>
          ) : (
            <button
              className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400"
              onClick={() => setSearchOpen(true)}
            >
              <Search size={20} />
            </button>
          )}
        </div>
        
        <div className="flex items-center space-x-3">
          <div className="relative">
            <select
              value={preset}
              onChange={(e) => setTimeRangeFromPreset(e.target.value as TimeRangePreset)}
              className="appearance-none w-40 px-4 py-2 border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {TIME_RANGE_PRESETS.map((range) => (
                <option key={range.value} value={range.value}>
                  {range.label}
                </option>
              ))}
              {preset === 'custom' && <option value="custom">Custom Range</option>}
            </select>
          </div>
          
          <button className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400 relative">
            <Bell size={20} />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
          </button>
          
          <button
            className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400"
            onClick={toggleTheme}
          >
            {theme.mode === 'light' ? <Moon size={20} /> : <Sun size={20} />}
          </button>
          
          <button className="p-2 rounded-md text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-800 dark:text-gray-400">
            <HelpCircle size={20} />
          </button>
        </div>
      </div>
    </header>
  );
};

export default Navbar;