import React, { createContext, useState, useContext, useEffect } from 'react';
import { ThemeOptions } from '../types/common';

interface ThemeContextType {
  theme: ThemeOptions;
  toggleTheme: () => void;
  setColorMode: (mode: 'light' | 'dark') => void;
  setPrimaryColor: (color: string) => void;
}

const defaultTheme: ThemeOptions = {
  mode: 'light',
  primaryColor: '#4285F4' // Google Blue
};

const ThemeContext = createContext<ThemeContextType>({
  theme: defaultTheme,
  toggleTheme: () => {},
  setColorMode: () => {},
  setPrimaryColor: () => {}
});

export const useTheme = () => useContext(ThemeContext);

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setTheme] = useState<ThemeOptions>(() => {
    // Initialize from localStorage if available
    const savedTheme = localStorage.getItem('theme');
    if (savedTheme) {
      try {
        return JSON.parse(savedTheme);
      } catch (e) {
        console.error('Failed to parse theme from localStorage:', e);
      }
    }
    
    // Default to system preference for color mode
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return { ...defaultTheme, mode: 'dark' };
    }
    
    return defaultTheme;
  });
  
  // Save theme to localStorage when it changes
  useEffect(() => {
    localStorage.setItem('theme', JSON.stringify(theme));
    
    // Apply theme to document
    if (theme.mode === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
    
    // Set CSS variables for primary color
    document.documentElement.style.setProperty('--color-primary', theme.primaryColor);
    
    // Generate lighter and darker variants of primary color for hover/focus states
    const hexToRgb = (hex: string): [number, number, number] => {
      const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
      return result 
        ? [
            parseInt(result[1], 16),
            parseInt(result[2], 16),
            parseInt(result[3], 16)
          ]
        : [0, 0, 0];
    };
    
    const [r, g, b] = hexToRgb(theme.primaryColor);
    
    // Lighten by 10%
    const lighten = (c: number) => Math.min(255, c + 25);
    const lighter = `rgb(${lighten(r)}, ${lighten(g)}, ${lighten(b)})`;
    
    // Darken by 10%
    const darken = (c: number) => Math.max(0, c - 25);
    const darker = `rgb(${darken(r)}, ${darken(g)}, ${darken(b)})`;
    
    document.documentElement.style.setProperty('--color-primary-light', lighter);
    document.documentElement.style.setProperty('--color-primary-dark', darker);
  }, [theme]);
  
  const toggleTheme = () => {
    setTheme(prevTheme => ({
      ...prevTheme,
      mode: prevTheme.mode === 'light' ? 'dark' : 'light'
    }));
  };
  
  const setColorMode = (mode: 'light' | 'dark') => {
    setTheme(prevTheme => ({
      ...prevTheme,
      mode
    }));
  };
  
  const setPrimaryColor = (primaryColor: string) => {
    setTheme(prevTheme => ({
      ...prevTheme,
      primaryColor
    }));
  };
  
  return (
    <ThemeContext.Provider value={{ theme, toggleTheme, setColorMode, setPrimaryColor }}>
      {children}
    </ThemeContext.Provider>
  );
};