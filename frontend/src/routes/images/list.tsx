import { useContext, useEffect, useState } from "react";
import { AxiosContext } from "../../context/axios";
import { PagedImages, } from "../../api/resources";
import { deleteImage, getImages } from "../../api/image";
import { Pages } from "../../components/pages";
import { ImageList } from "../../components/image_list";
import { PrimaryButton } from "../../components/button";
import { ImageUploader } from "../../components/image_uploader";

export const Images = () => {
    const axiosInstance = useContext(AxiosContext);
    const [images, setImages] = useState<PagedImages | undefined>(undefined);
    const [showUploader, setShowUploader] = useState(false);
    const [currentPage, setCurrentPage] = useState(0);
    const [uploadedKey, setUploadedKey] = useState(0);

    useEffect(() => {
        if (axiosInstance === null) {
            return;
        }
        getImages(axiosInstance, currentPage, 18)
            .then(setImages)
            .catch((reason) => {
                console.log(reason);
            });
    }, [axiosInstance, currentPage, uploadedKey]);
    
    const onUpload = () => {
        setUploadedKey((key) => key + 1);        
        setShowUploader(false);
    };
    
    const onDelete = (id: string) => {
        if (axiosInstance === null) {
            return;
        }
        const asyncDo = async () => {
            await deleteImage(axiosInstance, id);
            setUploadedKey((key) => key + 1);        
        };
        asyncDo();
    };

    if (!images) {
        return <p>Loading...</p>;
    }
    return <>
        {showUploader ? <PrimaryButton onClick={() => setShowUploader(false)}>Hide browser</PrimaryButton> : <PrimaryButton onClick={() => setShowUploader(true)}>Upload more images</PrimaryButton>}
        {showUploader ? <div className="mt-2"><ImageUploader onUpload={onUpload} /></div> : null}
        <ImageList images={images.images} onDelete={onDelete} />
        {images.totalCount === 0 ? <p className="mt-3 text-display">No images yet</p> : null}
        {images.totalCount > 0 &&
            <Pages
                currentPage={currentPage}
                onPageChange={(n) => setCurrentPage(n)}
                pages={images.totalPageCount}
            />
        }
    </>
}