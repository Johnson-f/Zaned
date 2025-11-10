# Frontend Integration Guide

This guide shows how to configure your frontend to connect to your HTTPS backend at `api.zaned.site`.

## Environment Configuration

### React / Next.js

Create `.env.production`:

```bash
# .env.production
NEXT_PUBLIC_API_URL=https://api.zaned.site
```

Create `.env.development`:

```bash
# .env.development
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Usage in code:

```javascript
// lib/api.js or utils/api.js
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export const fetchScreenerData = async () => {
  const response = await fetch(`${API_URL}/api/screener-results?type=inside_day`);
  return response.json();
};

// Or create an API client
export const api = {
  baseURL: API_URL,
  
  async get(endpoint) {
    const response = await fetch(`${this.baseURL}${endpoint}`);
    if (!response.ok) throw new Error('API request failed');
    return response.json();
  },
  
  async post(endpoint, data) {
    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    if (!response.ok) throw new Error('API request failed');
    return response.json();
  }
};

// Usage
const data = await api.get('/api/health');
```

### Vue / Nuxt

```javascript
// nuxt.config.js
export default {
  publicRuntimeConfig: {
    apiURL: process.env.API_URL || 'https://api.zaned.site'
  },
  privateRuntimeConfig: {
    apiURL: process.env.API_URL || 'https://api.zaned.site'
  }
}

// In components
export default {
  async asyncData({ $config }) {
    const response = await fetch(`${$config.apiURL}/api/health`);
    return { data: await response.json() };
  }
}
```

### Vite (React/Vue/Svelte)

```javascript
// .env.production
VITE_API_URL=https://api.zaned.site

// .env.development
VITE_API_URL=http://localhost:8080

// In your code
const API_URL = import.meta.env.VITE_API_URL;

fetch(`${API_URL}/api/health`)
  .then(res => res.json())
  .then(data => console.log(data));
```

### Angular

```typescript
// environment.prod.ts
export const environment = {
  production: true,
  apiUrl: 'https://api.zaned.site'
};

// environment.ts
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080'
};

// In service
import { environment } from '../environments/environment';

@Injectable()
export class ApiService {
  private apiUrl = environment.apiUrl;
  
  getHealth() {
    return this.http.get(`${this.apiUrl}/api/health`);
  }
}
```

---

## API Client Examples

### Axios Setup

```javascript
// api/client.js
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'https://api.zaned.site',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Request interceptor
apiClient.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

export default apiClient;

// Usage
import apiClient from './api/client';

const getScreenerResults = async (type, period) => {
  return apiClient.get('/api/screener-results', {
    params: { type, period }
  });
};
```

### Fetch Wrapper

```javascript
// api/fetch-wrapper.js
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'https://api.zaned.site';

class ApiClient {
  constructor(baseURL) {
    this.baseURL = baseURL;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'API request failed');
      }
      
      return response.json();
    } catch (error) {
      console.error('API Error:', error);
      throw error;
    }
  }

  get(endpoint, options) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  post(endpoint, data, options) {
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  put(endpoint, data, options) {
    return this.request(endpoint, {
      ...options,
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  delete(endpoint, options) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }
}

export const api = new ApiClient(API_URL);

// Usage
import { api } from './api/fetch-wrapper';

const data = await api.get('/api/screener-results?type=inside_day');
const result = await api.post('/api/watchlist', { name: 'My Watchlist' });
```

---

## React Query / SWR Integration

### React Query

```javascript
// api/queries.js
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from './client';

export const useScreenerResults = (type, period) => {
  return useQuery({
    queryKey: ['screener-results', type, period],
    queryFn: () => api.get(`/api/screener-results?type=${type}&period=${period}`),
  });
};

export const useMarketStats = () => {
  return useQuery({
    queryKey: ['market-stats'],
    queryFn: () => api.get('/api/market-statistics/live'),
    refetchInterval: 5 * 60 * 1000, // Refetch every 5 minutes
  });
};

// In component
function ScreenerResults() {
  const { data, isLoading, error } = useScreenerResults('inside_day', '7d');
  
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  
  return (
    <div>
      <h2>Results: {data.data.count}</h2>
      {/* Render results */}
    </div>
  );
}
```

### SWR

```javascript
// api/hooks.js
import useSWR from 'swr';

const fetcher = (url) => fetch(url).then(r => r.json());
const API_URL = process.env.NEXT_PUBLIC_API_URL;

export const useScreenerResults = (type, period) => {
  const { data, error, isLoading } = useSWR(
    `${API_URL}/api/screener-results?type=${type}&period=${period}`,
    fetcher
  );
  
  return {
    data: data?.data,
    isLoading,
    error
  };
};

export const useMarketStats = () => {
  const { data, error } = useSWR(
    `${API_URL}/api/market-statistics/live`,
    fetcher,
    { refreshInterval: 5 * 60 * 1000 } // Refresh every 5 minutes
  );
  
  return {
    stats: data?.data,
    error
  };
};
```

---

## Error Handling

### Global Error Handler

```javascript
// utils/error-handler.js
export class ApiError extends Error {
  constructor(message, status, data) {
    super(message);
    this.status = status;
    this.data = data;
  }
}

export const handleApiError = (error) => {
  if (error.response) {
    // Server responded with error
    const message = error.response.data?.message || 'Server error';
    throw new ApiError(message, error.response.status, error.response.data);
  } else if (error.request) {
    // Request made but no response
    throw new ApiError('No response from server', 0, null);
  } else {
    // Something else happened
    throw new ApiError(error.message, 0, null);
  }
};

// Usage
try {
  const data = await api.get('/api/screener-results');
} catch (error) {
  handleApiError(error);
}
```

---

## Testing API Connection

### Simple Test Script

```javascript
// test-api.js
const API_URL = 'https://api.zaned.site';

async function testAPI() {
  console.log('Testing API connection...');
  
  try {
    // Test health endpoint
    const healthResponse = await fetch(`${API_URL}/api/health`);
    const healthData = await healthResponse.json();
    console.log('✓ Health check:', healthData);
    
    // Test screener results
    const screenerResponse = await fetch(`${API_URL}/api/screener-results?type=inside_day&period=7d`);
    const screenerData = await screenerResponse.json();
    console.log('✓ Screener results:', screenerData.data.count, 'symbols');
    
    // Test market stats
    const statsResponse = await fetch(`${API_URL}/api/market-statistics/live`);
    const statsData = await statsResponse.json();
    console.log('✓ Market stats:', statsData.data);
    
    console.log('\n✓ All tests passed!');
  } catch (error) {
    console.error('✗ Test failed:', error.message);
  }
}

testAPI();
```

Run with: `node test-api.js`

---

## CORS Troubleshooting

If you get CORS errors:

1. **Check backend CORS configuration**:
   ```bash
   # In backend/.env
   ALLOWED_ORIGINS=https://zaned.site,https://www.zaned.site
   ```

2. **Restart backend**:
   ```bash
   docker-compose -f docker-compose.prod.yml restart backend
   ```

3. **Verify in browser console**:
   - Check the error message
   - Look for `Access-Control-Allow-Origin` header
   - Ensure you're using HTTPS (not HTTP)

4. **Test with curl**:
   ```bash
   curl -H "Origin: https://zaned.site" \
        -H "Access-Control-Request-Method: GET" \
        -X OPTIONS \
        https://api.zaned.site/api/health -v
   ```

---

## Production Checklist

Before deploying frontend:

- [ ] Update API URL to `https://api.zaned.site`
- [ ] Test API connection from frontend
- [ ] Verify CORS is configured correctly
- [ ] Check all API endpoints work
- [ ] Test error handling
- [ ] Verify authentication (if applicable)
- [ ] Test on production build
- [ ] Monitor for errors in production

---

## Example: Complete React App Setup

```javascript
// src/config/api.js
export const API_CONFIG = {
  baseURL: process.env.REACT_APP_API_URL || 'https://api.zaned.site',
  timeout: 10000,
};

// src/services/api.service.js
import axios from 'axios';
import { API_CONFIG } from '../config/api';

const apiClient = axios.create(API_CONFIG);

export const screenerService = {
  getResults: (type, period) => 
    apiClient.get('/api/screener-results', { params: { type, period } }),
  
  getMarketStats: () => 
    apiClient.get('/api/market-statistics/live'),
  
  getCompanyInfo: (symbol) => 
    apiClient.get(`/api/company-info/${symbol}`),
};

// src/App.js
import { useEffect, useState } from 'react';
import { screenerService } from './services/api.service';

function App() {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    screenerService.getResults('inside_day', '7d')
      .then(response => {
        setData(response.data);
        setLoading(false);
      })
      .catch(error => {
        console.error('Error:', error);
        setLoading(false);
      });
  }, []);

  if (loading) return <div>Loading...</div>;
  
  return (
    <div>
      <h1>Screener Results</h1>
      <p>Found {data?.data?.count} symbols</p>
    </div>
  );
}
```

---

## Need Help?

Common issues:

1. **CORS errors**: Update `ALLOWED_ORIGINS` in backend `.env`
2. **Connection refused**: Check if backend is running
3. **SSL errors**: Verify certificate is valid
4. **404 errors**: Check API endpoint paths

For more help, check backend logs:
```bash
docker-compose -f docker-compose.prod.yml logs backend
```
