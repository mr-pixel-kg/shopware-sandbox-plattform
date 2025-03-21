import ApiService from "@/services/apiService.js";

class ImagesService {

    async getAllImages() {
        return await ApiService.request("get", "/api/images");
    }

    async deleteImage(id) {
        return await ApiService.request("delete", `/api/images/${id}`);
    }

    async registerImage(imageName, imageTag) {
        return await ApiService.request("post", "/api/images", {
            image_name: imageName,
            image_tag: imageTag
        });
    }

}

export default new ImagesService();