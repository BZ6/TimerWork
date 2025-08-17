import axios from 'axios';

const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? '/api' 
  : 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Добавляем токен к каждому запросу
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Обработка ошибок авторизации
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.reload();
    }
    return Promise.reject(error);
  }
);

export const authAPI = {
  register: (username, password) =>
    api.post('/register', { username, password }),
  
  login: (username, password) =>
    api.post('/login', { username, password }),
};

export const timerAPI = {
  getWorkWeek: () => api.get('/workweek'),
  
  startWeek: (goalMinutes) => api.post('/workweek/start', { goal_minutes: goalMinutes }),
  
  endWeek: () => api.post('/workweek/end'),
  
  pauseTimer: () => api.post('/workweek/pause'),
  
  resumeTimer: () => api.post('/workweek/resume'),
  
  getCurrentTime: () => api.get('/workweek/current-time'),

  getWeekHistory: () => api.get('/workweek/history'),
};

export default api;
