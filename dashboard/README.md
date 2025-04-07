# Metrics Dashboard

A modern, responsive metrics dashboard built with React, TypeScript, and Tailwind CSS that offers a Google Analytics-like experience for monitoring cloud-native applications.

## Features

- **Dashboard**: Overview of system health with key metrics, charts, and alerts
- **Metrics Explorer**: Detailed analysis of individual metrics with historical data
- **Query Builder**: Create custom PromQL queries (coming soon)
- **Alerts Management**: View and manage alerts (coming soon)
- **Dark/Light Mode**: Support for both dark and light themes
- **Responsive Design**: Works well on desktop and mobile devices

## Technology Stack

- **Frontend**:
  - React 18 with TypeScript
  - Tailwind CSS for styling
  - Recharts for data visualization
  - React Router for navigation
  - Context API for state management

- **Backend**:
  - Connects to a Go API with Prometheus integration

## Getting Started

### Prerequisites

- Node.js 16+
- npm or yarn

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/prashanth8983/metrics-dashboard.git
   cd metrics-dashboard
   ```

2. Install dependencies:
   ```bash
   npm install
   # or
   yarn install
   ```

3. Configure the API endpoint:
   - Create a `.env` file in the root directory:
   ```
   REACT_APP_API_URL=http://localhost:8080/