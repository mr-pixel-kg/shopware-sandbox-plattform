import axios from "axios";

class ImagesService {

    constructor() {
        this.apiClient = axios.create({
            baseURL: "http://localhost:8080",
            headers: {
                "Content-Type": "application/json",
                "Access-Control-Allow-Origin": "*",
            },
        });
    }

    async getAllImages() {
        const response = await this.apiClient.get("/api/images");
        console.log(response)
        return response.data;
    }

    async deleteImage(id) {
        const response = await this.apiClient.delete(`/api/images/${id}`);
        return response.data;
    }

    async registerImage(imageName, imageTag) {
        var data = {
            "image_name": imageName,
            "image_tag": imageTag
        };
        const response = await this.apiClient.post("/api/images", data);
        return response.data;
    }

}

export default new ImagesService();