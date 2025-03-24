import ApiService from "@/services/apiService.js";
import { SandboxEnvironmentModel } from "@/models/SandboxEnvironmentModel.js";

class SandboxService {
  async getAllSandboxes() {
    return await ApiService.request("get", "/api/sandboxes");
  }

  async deleteSandbox(id) {
    try {
      const response = await ApiService.request(
        "delete",
        `/api/sandboxes/${id}`,
      );
      return { success: true, message: "Sandbox erfolgreich gelöscht" };
    } catch (e) {
      console.error("Failed to delete sandbox", e);
      const errorMessage =
        e.response?.data?.message ||
        e.message ||
        "Ein unbekannter Fehler ist aufgetreten";
      return { success: false, message: errorMessage };
    }
  }

  async createSandbox(image_name, lifetime) {
    try {
      const response = await ApiService.request("post", "/api/sandboxes", {
        image_name: image_name,
        lifetime: lifetime,
      });

      const sandbox = new SandboxEnvironmentModel(
        response.sandbox_id,
        response.image,
        "starting",
        Date.parse(response.created_at),
        Date.parse(response.destroy_at),
        response.url,
      );

      return { success: true, sandbox: sandbox };
    } catch (e) {
      console.log("Failed to create sandbox", e);
      const errorMessage =
        e.response?.data?.message ||
        e.message ||
        "Ein unbekannter Fehler ist aufgetreten";
      return { success: false, message: errorMessage };
    }
  }

  async refreshSandbox(sandbox) {
    try {
      const response = await ApiService.request(
        "GET",
        `/api/sandboxes/${sandbox.id}`,
      );

      sandbox.status = response.state; // Attention when refactor: state is not the same as status
      sandbox.createdAt = Date.parse(response.created_at);
      sandbox.destroyAt = Date.parse(response.destroy_at);

      return { success: true, sandbox: sandbox };
    } catch (e) {
      console.error("Failed to refresh sandbox details", e);
      return { success: false, sandbox: null };
    }
  }
}

export default new SandboxService();
