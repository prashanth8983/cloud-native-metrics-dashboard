import axios from 'axios';

// Create Axios instance with default config
const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Get token from localStorage
    const token = localStorage.getItem('authToken');
    
    // If token exists, add to headers
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    // Handle generic errors (like network issues)
    if (!error.response) {
      console.error('Network Error:', error);
      return Promise.reject(new Error('Network error. Please check your connection.'));
    }
    
    // Handle API errors
    const { status, data } = error.response;
    
    switch (status) {
      case 401:
        // Handle unauthorized
        localStorage.removeItem('authToken');
        window.location.href = '/login';
        break;
      case 403:
        // Handle forbidden
        console.error('Forbidden access:', data);
        break;
      case 404:
        // Handle not found
        console.error('Resource not found:', data);
        break;
      case 500:
        // Handle server error
        console.error('Server error:', data);
        break;
      default:
        // Handle other errors
        console.error(`Error ${status}:`, data);
    }
    
    return Promise.reject(error);
  }
);

export default api;