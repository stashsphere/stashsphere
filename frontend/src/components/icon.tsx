type IconProps = {
  icon: string;
  className?: string;
  height?: string;
  width?: string;
};

export const Icon = ({ icon, height, width, className }: IconProps) => {
  const style: React.CSSProperties = {};
  if (height !== undefined) {
    style['height'] = height;
    style['width'] = width;
  }
  return <span className={'iconify ' + icon + ' ' + (className || '')} style={style}></span>;
};
