import { Axios } from "axios";
import { PagedImages, ReducedImage } from "./resources";

export const createImage = async (axios: Axios, image: File): Promise<ReducedImage> => {
    const data = new FormData();
    data.set("file", image);
    const response = await axios.post(`/images`, data, {
        headers: {'Content-Type': 'multipart/form-data' }
    });

    return response.data as ReducedImage;
}

export const fetchImageData = async (axios: Axios, id: string): Promise<File> => {
    const response = await axios.get(`/images/${id}`);
    return response.data as File;
}

export const getImages = async (axios: Axios, currentPage: number, perPage: number) => {
    const response = await axios.get(`/images?page=${currentPage}&perPage=${perPage}`, { headers: {
            "Content-Type": "application/json"
    }});
    
    if (response.status != 200) {
        throw `Got error ${response}`
    }
   
    const images = response.data as PagedImages;
    return images;
}

export const deleteImage = async (axios: Axios, id: string): Promise<ReducedImage> => {
    const response = await axios.delete(`/images/${id}`);
    return response.data as ReducedImage;
}