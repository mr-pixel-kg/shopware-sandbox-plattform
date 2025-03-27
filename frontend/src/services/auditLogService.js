import ApiService from "@/services/apiService.js";

class AuditLogService {
  async getAllLogEntries() {
    return await ApiService.request("get", "/api/auditlog");
  }
}

export default new AuditLogService();
