import { Property } from "../api/resources";
import { Icon } from "./icon";

export type PropertyProps = {
    property: Property,
    keyWidth: string,
}

// these colors are defined by SVG and also by CSS
const definedSVGColors = [
     "aliceblue", "antiquewhite", "aqua", "aquamarine", "azure", "beige", "bisque", "black", "blanchedalmond", "blue", "blueviolet", "brown", "burlywood", "cadetblue", "chartreuse", "chocolate", "coral", "cornflowerblue", "cornsilk", "crimson", "cyan", "darkblue", "darkcyan", "darkgoldenrod", "darkgray", "darkgreen", "darkgrey", "darkkhaki", "darkmagenta", "darkolivegreen", "darkorange", "darkorchid", "darkred", "darksalmon", "darkseagreen", "darkslateblue", "darkslategray", "darkslategrey", "darkturquoise", "darkviolet", "deeppink", "deepskyblue", "dimgray", "dimgrey", "dodgerblue", "firebrick", "floralwhite", "forestgreen", "fuchsia", "gainsboro", "ghostwhite", "gold", "goldenrod", "gray", "green", "greenyellow", "grey", "honeydew", "hotpink", "indianred", "indigo", "ivory", "khaki", "lavender", "lavenderblush", "lawngreen", "lemonchiffon", "lightblue", "lightcoral", "lightcyan", "lightgoldenrodyellow", "lightgray", "lightgreen", "lightgrey", "lightpink", "lightsalmon", "lightseagreen", "lightskyblue", "lightslategray", "lightslategrey", "lightsteelblue", "lightyellow", "lime", "limegreen", "linen", "magenta", "maroon", "mediumaquamarine", "mediumblue", "mediumorchid", "mediumpurple", "mediumseagreen", "mediumslateblue", "mediumspringgreen", "mediumturquoise", "mediumvioletred", "midnightblue", "mintcream", "mistyrose", "moccasin", "navajowhite", "navy", "oldlace", "olive", "olivedrab", "orange", "orangered", "orchid", "palegoldenrod", "palegreen", "paleturquoise", "palevioletred", "papayawhip", "peachpuff", "peru", "pink", "plum", "powderblue", "purple", "rebeccapurple", "red", "rosybrown", "royalblue", "saddlebrown", "salmon", "sandybrown", "seagreen", "seashell", "sienna", "silver", "skyblue", "slateblue", "slategray", "slategrey", "snow", "springgreen", "steelblue", "tan", "teal", "thistle", "tomato", "transparent", "turquoise", "violet", "wheat", "white", "whitesmoke", "yellow", "yellowgreen"
   ];

const formatColor = (color: string | number) => {
    const asString = color.toString();
    if (definedSVGColors.includes(asString)) {
        const elementStyle = {
            backgroundColor: asString,
         };
        return  <div className="flex flex-row gap-1">
                    <span>{color}</span>
                    <div style={elementStyle} className="ml-2 w-4 h-4 border border-black rounded"></div>
                </div>
    } else {
        return <>{color}</>;
    }
}

const isValidURI = (uri: string): boolean => {
    try { 
        new URL(uri);
        return true;
     } catch (_) {
        return false;
     }
}

const formatValue = (value: string | number) => {
    const asString = value.toString();
    if (isValidURI(asString)) {
        return <a href={asString} target="_blank" rel="noopener noreferrer">{asString}</a>;
    } else {
        return <>{asString}</>;
    }
}

const PropertyComponent = ({ property, keyWidth }: PropertyProps) => {
    // help tailwind to generate all widths used
    // min-w-[1rem] min-w-[2rem] min-w-[3rem] min-w-[4rem] min-w-[5rem] min-w-[6rem]
    // min-w-[7rem] min-w-[8rem] min-w-[9rem] min-w-[10rem] min-w-[11rem] min-w-[12rem]
    // min-w-[13rem] min-w-[14rem] min-w-[15rem] min-w-[16rem] min-w-[17rem] min-w-[18rem] min-w-[19rem]
    const [first, second] = (() => {
        switch (property.name) {
            case "manufacturer":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="manufacturer">
                        <Icon icon={"mdi--manufacturing"} />
                        Manufacturer:
                    </div>,
                    property.value
                ];
            case "manufacturer_url":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="manufacturer_url">
                        <Icon icon={"mdi--link-variant"} />
                        Manufacturer Link:
                    </div>,
                    formatValue(property.value)
                ];
            case "torque_min":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="torque_min">
                        <Icon icon={"mdi--refresh"} />
                        Min. Torque:
                    </div>,
                    property.value
                ];
            case "torque_max":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="torque_max">
                        <Icon icon={"mdi--refresh"} />
                        Max. Torque:
                    </div>,
                    property.value
                ];
            case "material":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="material">
                        <Icon icon={"mdi--test-tube"} />
                        Material:
                    </div>,
                    property.value
                ];
            case "color":
                return [
                    <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key="color">
                        <Icon icon={"mdi--color"} />
                        Color:
                    </div>,
                    formatColor(property.value)
                ];
            default:
                switch (property.type) {
                    case "string":
                        return [
                            <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key={property.name}>
                                <Icon icon="mdi--format-text" />
                                {property.name}:
                            </div>,
                            formatValue(property.value)
                        ];
                    case "datetime":
                        return [
                            <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key={property.name}>
                                <Icon icon="mdi--date-range" />
                                {property.name}:
                            </div>,
                            property.value.toString()
                        ];
                    case "float":
                        return [
                            <div className={`flex flex-nowrap min-w-[${keyWidth}]`} key={property.name}>
                                <Icon icon="mdi--hashtag" />
                                {property.name}:
                            </div>,
                            property.value
                        ];
                }
        }
    })();

    if (property.type === "float") {
        return (
            <div className="flex flex-row items-center gap-2">
                <span className="font-semibold whitespace-nowrap">{first}</span>
                <span>{second} {property.unit}</span>
            </div>
        );
    } else {
        return (
            <div className="flex flex-row items-center gap-2">
                <span className="font-semibold whitespace-nowrap">{first}</span>
                <span>{second}</span>
            </div>
        );
    }
};

export default PropertyComponent;