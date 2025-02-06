type IconProps = {
    icon: string
    height?: string
    width?: string
};

export const Icon = ({icon, height, width}: IconProps) => {
    const style: React.CSSProperties = {};
    if (height !== undefined) {
        style["height"] = height;
        style["width"] = width;
    }
    return <span className={"iconify " + icon} style={style}></span>
}