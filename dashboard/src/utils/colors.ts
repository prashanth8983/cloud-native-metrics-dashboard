/**
 * Color palette for charts consistent with Material Design
 */
export const chartColors = [
    '#4285F4', // Google Blue
    '#DB4437', // Google Red
    '#F4B400', // Google Yellow
    '#0F9D58', // Google Green
    '#AB47BC', // Purple
    '#00ACC1', // Cyan
    '#FF7043', // Deep Orange
    '#9E9E9E', // Grey
    '#5C6BC0', // Indigo
    '#3949AB', // Dark Blue
  ];
  
  /**
   * Get a color for severity level
   */
  export function getSeverityColor(severity: string): string {
    switch (severity.toLowerCase()) {
      case 'critical':
        return '#DB4437'; // Red
      case 'high':
      case 'warning':
        return '#F4B400'; // Yellow
      case 'medium':
        return '#FF7043'; // Orange
      case 'low':
        return '#0F9D58'; // Green
      case 'info':
        return '#4285F4'; // Blue
      default:
        return '#9E9E9E'; // Grey
    }
  }
  
  /**
   * Generate a color based on a string (consistent hashing)
   */
  export function stringToColor(str: string): string {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      hash = str.charCodeAt(i) + ((hash << 5) - hash);
    }
    
    let color = '#';
    for (let i = 0; i < 3; i++) {
      const value = (hash >> (i * 8)) & 0xFF;
      color += ('00' + value.toString(16)).substr(-2);
    }
    
    return color;
  }
  
  /**
   * Get color for status
   */
  export function getStatusColor(status: string): string {
    switch (status.toLowerCase()) {
      case 'up':
      case 'success':
      case 'healthy':
      case 'normal':
        return '#0F9D58'; // Green
      case 'degraded':
      case 'warning':
        return '#F4B400'; // Yellow
      case 'down':
      case 'error':
      case 'critical':
        return '#DB4437'; // Red
      default:
        return '#9E9E9E'; // Grey
    }
  }