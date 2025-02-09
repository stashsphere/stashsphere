import { useContext } from "react";
import { ConfigContext } from "../context/config";

type ImageGridProps = {
    imageIds: string[];
};


const ImageGrid = (props: ImageGridProps) => {
    const config = useContext(ConfigContext);
    const imageIds = props.imageIds.slice(0, props.imageIds.length > 4 ? 4 : props.imageIds.length + 1);

    return (
        <div className="grid grid-cols-2 gap-2 m-4">
            {imageIds.map(hash =>
                <div className="flex w-30 h-30 items-center justify-center rounded-md">
                    <img src={`${config.apiHost}/api/images/${hash}`} alt="Image" className="object-contain w-full h-full" />
                </div>
            )}
        </div>
    )
}


export default ImageGrid;