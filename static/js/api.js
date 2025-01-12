// api.js
class ApiService {
    constructor(baseUrl) {
        this.baseUrl = baseUrl;
    }

    async request(endpoint, options = {}) {
        const defaultOptions = {
            credentials: 'include', // Важно за cookies
            headers: {
                'Accept': 'application/json',
                'Content-Type': 'application/json',
                'Cache-Control': 'no-cache',
                'Pragma': 'no-cache'
            },
            mode: 'same-origin'
        };

        const config = {
            ...defaultOptions,
            ...options,
            headers: {
                ...defaultOptions.headers,
                ...options.headers
            }
        };
        console.log('request method called from:', new Error().stack);
        try {
            const response = await fetch(`${this.baseUrl}${endpoint}`, config);

            // Проверка за автентикация
            if (response.status === 401) {
                // Можете да добавите логика за refresh token или redirect
                window.location.href = '/#login';
                return;
            }

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    }

    // Auth методи
    async login(credentials) {
        return this.request(`${config.endpoints.login}`, {
            method: 'POST',
            body: JSON.stringify(credentials)
        });
    }

    async logout() {
        return this.request(`${config.endpoints.logout}`, {
            method: 'POST'
        });
    }

    async checkAuth() {
        console.log('checkAuth method called from:', new Error().stack);
        return this.request(`${config.endpoints.authStatus}`);
    }

    // CRUD операции
    async get(endpoint) {
        return this.request(endpoint);
    }

    async post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async delete(endpoint) {
        return this.request(endpoint, {
            method: 'DELETE'
        });
    }
}

// Създаване на инстанция
const api = new ApiService(window.location.origin);

// export default api;