import { ChangeEvent, useContext, useMemo, useRef, useState } from "react";
import { DangerButton, PrimaryButton } from "./button"
import { Icon } from "./icon";
import { createImage } from "../api/image";
import { AxiosContext } from "../context/axios";
import { ReducedImage } from "../api/resources";

type ImageUploaderProps = {
    onUpload: (images: ReducedImage[]) => void;
}

export const ImageUploader = ({onUpload}: ImageUploaderProps) => {
    const [files, setFiles] = useState<File[]>([]);
    const axiosInstance = useContext(AxiosContext);

    const previewUrls = useMemo(() => {
        const urls = [];

        for (const file of files) {
            urls.push(URL.createObjectURL(file));
        }
        return urls;
    }, [files]);

    const onFileChange = (e: ChangeEvent<HTMLInputElement>) => {
        const v = [];
        for (let i = 0; i < (e.target.files?.length || 0); i++) {
            const item = e.target.files?.item(i);
            if (item) {
                v.push(item);
            }
        }
        setFiles([...files, ...v]);
        e.target.value = "";
    };

    const removeFile = (idx: number) => {
        const newFiles = files.filter((_, i) => i !== idx);
        setFiles(newFiles);
    };
    
    const onUploadClick = async () => {
        if (axiosInstance === null) {
            return
        }
        if (files.length === 0) {
            return
        }
        const uploadedImages = [];
        for (const file of files) {
            const image = await createImage(axiosInstance, file);
            uploadedImages.push(image);
        }
        onUpload(uploadedImages);
    };

    const inputRef = useRef<HTMLInputElement>(null);

    return (
        <div className="bg-neutral p-2 rounded">
            <h2 className="text-xl font-bold mb-4 text-onneutral">Image Uploader</h2>
            <div className="mb-4">
                <div className="flex flex-wrap gap-4">
                    {previewUrls.map((url, idx) => (
                        <div key={url}>
                            <div className="flex items-center gap-4 mb-2 flex-col">
                                <div className="flex w-60 h-60 items-center justify-center rounded-md">
                                    <img className="w-full h-full object-contain" src={url} alt="Preview" />
                                </div>
                                <DangerButton onClick={() => removeFile(idx)}>
                                    <Icon icon="mdi--trash" />
                                    Remove
                                </DangerButton>
                            </div>
                        </div>
                    ))}
                </div>
                <input ref={inputRef} type="file" accept="image/*" onChange={onFileChange} multiple hidden />
                <div className="flex flex-row gap-4">
                    <PrimaryButton onClick={() => inputRef.current?.click()}>Browse</PrimaryButton>
                    <PrimaryButton onClick={onUploadClick} disabled={ files.length === 0 ? true : false }>Upload</PrimaryButton>
                </div>
            </div>
        </div>);
}