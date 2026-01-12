// API Configuration
const API_BASE_URL = "http://localhost:8080/api";

const API_ENDPOINTS = {
  // Auth endpoints
  USER_SIGNUP: `${API_BASE_URL}/auth/signup`,
  USER_LOGIN: `${API_BASE_URL}/auth/login`,
  ADMIN_SIGNUP: `${API_BASE_URL}/auth/create-admin`,
  ADMIN_LOGIN: `${API_BASE_URL}/auth/login`, // Admin uses same login endpoint
};

// Store token in localStorage
const TOKEN_KEY = "disney_auth_token";
const USER_KEY = "disney_user_data";

// Helper functions
function saveToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

function saveUser(userData) {
  localStorage.setItem(USER_KEY, JSON.stringify(userData));
}

function getUser() {
  const userData = localStorage.getItem(USER_KEY);
  return userData ? JSON.parse(userData) : null;
}

function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

function isLoggedIn() {
  return !!getToken();
}
