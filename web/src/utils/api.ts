import axios from "axios";

export const getAuthToken = (): string => {
  return localStorage.getItem("mcp_auth_token") || "";
};

export const setAuthToken = (token: string) => {
  if (token) {
    localStorage.setItem("mcp_auth_token", token);
  } else {
    localStorage.removeItem("mcp_auth_token");
  }
};

const api = axios.create({
  baseURL: "",
});

api.interceptors.request.use((config) => {
  const token = getAuthToken();
  if (token) {
    config.headers["Authorization"] = `Bearer ${token}`;
  }
  return config;
});

export default api;
