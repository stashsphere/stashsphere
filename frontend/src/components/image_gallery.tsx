import { useContext, useState } from "react";
import { ReducedImage } from "../api/resources";
import { ConfigContext } from "../context/config";

type ImageGalleryProps = {
  images: ReducedImage[];
};

export const ImageGallery = ({ images }: ImageGalleryProps) => {
  const [selectedImage, setSelectedImage] = useState<ReducedImage | null>(images.length === 0 ? null : images[0]);
  const config = useContext(ConfigContext);

  const urlForImage = (image: ReducedImage) => {
    return `${config.apiHost}/api/images/${image.id}`;
  };

  // Determine the class for the selected thumbnail
  const getThumbnailClass = (image: ReducedImage) => {
    return image === selectedImage ? "border-4 border-blue-500" : "";
  };

  return (
    selectedImage ?(
    <div className="w-full relative">
      <img
        src={urlForImage(selectedImage)}
        alt="Main"
        className="w-full h-auto mb-4 rounded"
      />
      <div className="flex justify-around">
        {images.map((image, index) => (
          <img
            key={index}
            src={urlForImage(image)}
            alt="Thumbnail"
            onClick={() => setSelectedImage(image)}
            className={`w-24 h-auto cursor-pointer rounded shadow hover:shadow-lg transition duration-300 ease-in-out transform hover:scale-105 ${getThumbnailClass(image)}`}
          />
        ))}
      </div>
    </div>) : null
  );
};
