import axios from "axios";
import { useAuthStore } from "@/stores/authStore";

class ApiService {
  constructor() {
    this.apiClient = axios.create({
      baseURL: import.meta.env.VITE_BACKEND_URL || "http://localhost:8080",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
      },
    });
  }

  async login(username, password) {
    try {
      const response = await this.apiClient.get("/api/auth", {
        auth: { username, password },
      });

      console.log(response);

      return response.data.loggedIn === "true";
    } catch (error) {
      console.error("Login failed:", error);
      return false;
    }
  }

  logout() {
    const authStore = useAuthStore();
    return authStore.isAuthenticated;
  }

  isLoggedIn() {
    return this.authCredentials !== null;
  }

  async request(method, url, data = null) {
    const authStore = useAuthStore();
    const config = {
      method: method,
      url: url,
      data: data,
      auth: authStore.isAuthenticated
        ? { username: authStore.username, password: authStore.password }
        : undefined,
    };

    try {
      const response = await this.apiClient.request(config);
      console.log("API Response received: ", response);
      return response.data;
    } catch (error) {
      console.error("API Request Error:", error);
      throw error;
    }
  }
}

export default new ApiService();
