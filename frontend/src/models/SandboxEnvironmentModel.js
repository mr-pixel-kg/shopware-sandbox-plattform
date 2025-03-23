export class SandboxEnvironmentModel {
    constructor(id, imageName, status, createdAt, destroyAt, sandboxUrl) {
        this.id = id;
        this.imageName = imageName;
        this.status = status;
        this.createdAt = createdAt
        this.destroyAt = destroyAt;
        this.sandboxUrl = sandboxUrl;
    }

    // Optional: Methode zur Berechnung der verbleibenden Zeit
    getRemainingTime() {
        const now = new Date();
        const diffMs = this.destroyAt - now;
        const minutes = Math.floor(diffMs / (1000 * 60));
        const seconds = Math.floor((diffMs % (1000 * 60)) / 1000);
        return `${minutes}m ${seconds}s`;
    }

    getStorefrontUrl() {
        return this.sandboxUrl;
    }

    getAdminUrl() {
        return this.sandboxUrl + '/admin';
    }
}