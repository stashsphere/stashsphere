import { useState } from "react";
import { ReducedImage } from "../api/resources";
import { ImageComponent } from "./image";

type ImageGalleryProps = {
  images: ReducedImage[];
};

export const ImageGallery = ({ images }: ImageGalleryProps) => {
  const [selectedImage, setSelectedImage] = useState<ReducedImage | null>(images.length === 0 ? null : images[0]);

  // Determine the class for the selected thumbnail
  const getThumbnailClass = (image: ReducedImage) => {
    return image === selectedImage ? "border-4 border-blue-500" : "";
  };

  return (
    selectedImage ?(
    <div className="w-full relative">
      <ImageComponent image={selectedImage} defaultWidth={1024} className="w-full h-auto mb-4 rounded" alt="Main" />
      <div className="flex justify-around">
        {images.map((image, index) => (
          <ImageComponent image={image} defaultWidth={300} alt="Thumbnail" 
            key={index}
            onClick={() => setSelectedImage(image)}
            className={`w-24 h-auto cursor-pointer rounded shadow hover:shadow-lg transition duration-300 ease-in-out transform hover:scale-105 ${getThumbnailClass(image)}`}
          />
        ))}
      </div>
    </div>) : null
  );
};
