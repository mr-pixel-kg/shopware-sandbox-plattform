import ApiService from "@/services/apiService.js";

class SandboxService {

    async getAllSandboxes() {
        return await ApiService.request("get", "/api/sandboxes");
    }

    async deleteSandbox(id) {
        return await ApiService.request("delete", `/api/sandboxes/${id}`);
    }

    async createSandbox(image_name, lifetime) {
        return await ApiService.request("post", "/api/sandboxes", {
            image_name: image_name,
            lifetime: lifetime
        });
    }

}

export default new SandboxService();