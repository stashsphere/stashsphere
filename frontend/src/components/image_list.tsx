import { useContext, useMemo } from "react";
import { Image } from "../api/resources";
import { ConfigContext } from "../context/config";
import { AccentButton, DangerButton, PrimaryButton } from "./button";

type InterActionProps = {
    onDelete?: (id: string) => void;
    onSelect?: (id: string, selected: boolean) => void;
};

type ImageTileProps = {
    image: Image;
    selected?: boolean;
} & InterActionProps;

export const ImageTile = ({ image, onDelete, onSelect, selected }: ImageTileProps) => {
    const config = useContext(ConfigContext);

    const usedText = useMemo(() => {
        switch (image.things.length) {
            case 0:
                return "Used by no things";
            case 1:
                return "Used by one thing";
            case 2:
                return "Used by two things";
            default:
                return `Used by ${image.things.length} things`;
        }
    }, [image]);

    return <div className="flex w-full h-60 items-center justify-center rounded-md border border-secondary p-1 flex-col">
        <img src={`${config.apiHost}/api/images/${image.id}`} alt="Image" className="object-contain w-full h-full" />
        <span className="text-display">Name: {image.name}</span>
        <span className="text-display">{usedText}</span>
        {onDelete && image.actions.canDelete && <DangerButton onClick={() => onDelete(image.id)}>Delete</DangerButton>}
        {(onSelect !== undefined && selected !== undefined) &&
            (selected ? <AccentButton onClick={() => onSelect(image.id, false)}>Unselect</AccentButton> :
                <PrimaryButton onClick={() => onSelect(image.id, true)}>Select</PrimaryButton>)}
    </div>

}

type ImageListProps = {
    images: Image[];
    selectedImageIds?: string[];
} & InterActionProps;

export const ImageList = ({ images, selectedImageIds, ...rest }: ImageListProps) => {
    return <div className="grid grid-cols-3 gap-2 m-4">
        {images.map(image =>
            <ImageTile key={image.id} image={image} {...rest} selected={selectedImageIds?.includes(image.id)} />
        )}
    </div>
}