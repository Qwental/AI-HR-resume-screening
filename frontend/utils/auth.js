// utils/auth.js
export const setToken = (token) => {
    if (typeof window !== 'undefined') {
        localStorage.setItem('token', token);
    }
};

export const getToken = () => {
    if (typeof window !== 'undefined') {
        return localStorage.getItem('token');
    }
    return null;
};

export const removeToken = () => {
    if (typeof window !== 'undefined') {
        localStorage.removeItem('token');
    }
};

export const isAuthenticated = () => {
    return !!getToken();
};

// Функция для API запросов с автоматическим добавлением токена
// utils/auth.js
export const apiRequest = async (url, options = {}) => {
    const token = getToken();

    // Определяем базовый URL в зависимости от роута
    let baseUrl = '';
    if (url.startsWith('/api/interview/')) {
        baseUrl = process.env.NEXT_PUBLIC_INTERVIEW_API_URL || 'http://localhost:8081';
    } else if (url.startsWith('/api/auth/')) {
        baseUrl = process.env.NEXT_PUBLIC_AUTH_API_URL || 'http://localhost:8080';
    }

    const config = {
        ...options,
        headers: {
            'Content-Type': options.body instanceof FormData ? undefined : 'application/json',
            ...options.headers,
        },
    };

    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(baseUrl + url, config);

    if (response.status === 401) {
        removeToken();
        window.location.href = '/login';
    }

    return response;
};
