import axios from "axios";

class ApiService {
  constructor() {
    this.apiClient = axios.create({
      baseURL: import.meta.env.VITE_BACKEND_URL || "http://localhost:8080",
      headers: {
        "Content-Type": "application/json",
        "Access-Control-Allow-Origin": "*",
      },
    });
    this.authCredentials = null;
  }

  async login(username, password) {
    const response = await this.apiClient.get("/api/auth", {
      auth: {
        username: username,
        password: password,
      },
    });
    console.log(response);

    if (response.data.loggedIn === "true") {
      this.authCredentials = { username, password };
      return true;
    }

    return false;
  }

  logout() {
    this.authCredentials = null;
  }

  isLoggedIn() {
    return this.authCredentials !== null;
  }

  async request(method, url, data = null) {
    const config = {
      method: method,
      url: url,
      data: data,
      auth: this.authCredentials ? this.authCredentials : undefined,
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
