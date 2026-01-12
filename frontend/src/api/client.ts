import axios from 'axios';

const apiClient = axios.create({
  // Using a relative path allows Caddy to handle the routing.
  baseURL: '/api', 
  timeout: 5000, // Disconnect if the server doesn't respond in 5 seconds.
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
});

// You can add global error handling here (e.g., logging 500 errors).
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export default apiClient;