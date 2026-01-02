import { useContext, useMemo } from 'react';
import { Image, ReducedImage } from '../../api/resources';
import { ConfigContext } from '../../context/config';
import { urlForImage } from '../../api/image';

interface ImageComponentProps extends React.ImgHTMLAttributes<HTMLImageElement> {
  defaultWidth: number;
  image: Image | ReducedImage;
}

export const ImageComponent = ({ image, defaultWidth, ...rest }: ImageComponentProps) => {
  const config = useContext(ConfigContext);

  const [mainImage, srcSet, sizes] = useMemo(() => {
    const widths = [0.5, 1, 1.5, 2, 3].map((x) => x * defaultWidth);
    const images = widths.map((width) => urlForImage(config, image.hash, width));
    const srcSet = images.map((url, idx) => `${url} ${widths[idx]}w`).join(', ');
    const mainImage = images[1];
    const sizes = widths.map((x) => `${x}px`).join(', ');

    return [mainImage, srcSet, sizes];
  }, [config, image, defaultWidth]);

  return <img src={mainImage} sizes={sizes} srcSet={srcSet} {...rest} />;
};
