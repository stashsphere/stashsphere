import { useContext, useEffect, useMemo, useState } from "react";
import { AxiosContext } from "../context/axios";
import { PagedImages, Image } from "../api/resources";
import { getImages } from "../api/image";
import { ImageList } from "./image_list";
import { Pages } from "./pages";

export interface ImageBrowserProps {
    onSelected: (image: Image[]) => void;
}

export const ImageBrowser = ({onSelected}: ImageBrowserProps) => {
    const axiosInstance = useContext(AxiosContext);
    const [currentPage, setCurrentPage] = useState(0);
    const [images, setImages] = useState<PagedImages | undefined>(undefined);
    
    const [selectedImages, setSelectedImages] = useState<Image[]>([]);

    useEffect(() => {
        if (axiosInstance === null) {
            return;
        }
        getImages(axiosInstance, currentPage, 18)
            .then(setImages)
            .catch((reason) => {
                console.log(reason);
            });
    }, [axiosInstance, currentPage]);

    const selectedImageIds = useMemo(() => {
        return selectedImages.map(e => e.id);
    }, [selectedImages])
    
    useEffect(() => {
        onSelected(selectedImages);
    }, [onSelected, selectedImages]);
    
    const onSelect = (id: string, selected: boolean) => {
        console.log("called", id, selected)
        const selectedImage = images?.images.find((e) => e.id === id);
        if (!selectedImage) {
            return;
        }
        const newSelectedImages = selected ? [...selectedImages,  selectedImage] : selectedImages.filter((e) => e.id!== id);
        setSelectedImages(newSelectedImages);
    };
    
    if (!images) {
        return <p>Loading...</p>;
    }

    return <><ImageList images={images.images} selectedImageIds={selectedImageIds} onSelect={onSelect} />
        {images.totalCount > 0 &&
            <Pages
                currentPage={currentPage}
                onPageChange={(n) => setCurrentPage(n)}
                pages={images.totalPageCount}
            />
        }</>;
};