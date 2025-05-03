type IconProps = {
  icon: string;
  size?: 'small' | 'medium' | 'large';
  className?: string;
  tooltip?: string;
};

export const Icon = ({ icon, className, tooltip, size }: IconProps) => {
  const style: React.CSSProperties = {};
  // TODO: replace with CSS vars from theme
  switch (size) {
    case 'small':
      style['height'] = '16px';
      style['width'] = '16px';
      break;
    case 'medium':
    default:
      style['height'] = '24px';
      style['width'] = '24px';
      break;
    case 'large':
      style['height'] = '32px';
      style['width'] = '32px';
      break;
  }

  return (
    <span
      className={'iconify ' + icon + ' ' + (className || '')}
      style={style}
      title={tooltip}
    ></span>
  );
};
