import { Property } from "../api/resources";
import { Icon } from "./icon";

export type PropertyProps = {
    property: Property,
}

// these colors are defined by SVG and also by CSS
const definedSVGColors = [
    "aliceblue", "antiquewhite", "aqua", "aquamarine", "azure", "beige", "bisque", "black", "blanchedalmond", "blue", "blueviolet", "brown", "burlywood", "cadetblue", "chartreuse", "chocolate", "coral", "cornflowerblue", "cornsilk", "crimson", "cyan", "darkblue", "darkcyan", "darkgoldenrod", "darkgray", "darkgreen", "darkgrey", "darkkhaki", "darkmagenta", "darkolivegreen", "darkorange", "darkorchid", "darkred", "darksalmon", "darkseagreen", "darkslateblue", "darkslategray", "darkslategrey", "darkturquoise", "darkviolet", "deeppink", "deepskyblue", "dimgray", "dimgrey", "dodgerblue", "firebrick", "floralwhite", "forestgreen", "fuchsia", "gainsboro", "ghostwhite", "gold", "goldenrod", "gray", "green", "greenyellow", "grey", "honeydew", "hotpink", "indianred", "indigo", "ivory", "khaki", "lavender", "lavenderblush", "lawngreen", "lemonchiffon", "lightblue", "lightcoral", "lightcyan", "lightgoldenrodyellow", "lightgray", "lightgreen", "lightgrey", "lightpink", "lightsalmon", "lightseagreen", "lightskyblue", "lightslategray", "lightslategrey", "lightsteelblue", "lightyellow", "lime", "limegreen", "linen", "magenta", "maroon", "mediumaquamarine", "mediumblue", "mediumorchid", "mediumpurple", "mediumseagreen", "mediumslateblue", "mediumspringgreen", "mediumturquoise", "mediumvioletred", "midnightblue", "mintcream", "mistyrose", "moccasin", "navajowhite", "navy", "oldlace", "olive", "olivedrab", "orange", "orangered", "orchid", "palegoldenrod", "palegreen", "paleturquoise", "palevioletred", "papayawhip", "peachpuff", "peru", "pink", "plum", "powderblue", "purple", "rebeccapurple", "red", "rosybrown", "royalblue", "saddlebrown", "salmon", "sandybrown", "seagreen", "seashell", "sienna", "silver", "skyblue", "slateblue", "slategray", "slategrey", "snow", "springgreen", "steelblue", "tan", "teal", "thistle", "tomato", "transparent", "turquoise", "violet", "wheat", "white", "whitesmoke", "yellow", "yellowgreen"
  ];

const formatColor = (color: string) => {
    if (definedSVGColors.includes(color)) {
        const elementStyle = {
            backgroundColor: color,
        };
        return <div className="flex flex-row gap-1">red <div style={elementStyle} className="ml-2 w-4 h-4 border border-black rounded"></div></div>
    } else {
        return <>{color}</>
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
        const link = <a href={asString} target="_blank"><span className="">{asString}</span></a>
        return link
    } else {
        return <>{asString}</>
    }
}

const PropertyComponent = ({ property }: PropertyProps) => {
    const [first, second] = (() => {
        switch (property.name) {
            case "manufacturer":
                return [<span><Icon icon={"mdi--manufacturing"} />Manufacturer</span>, property.value]
            case "manufacturer_url":
                return [<span><Icon icon={"mdi--manufacturing"} />Manufacturer Link</span>, formatValue(property.value)]
            case "torque_min":
                return [<span><Icon icon={"mdi--refresh"} />Min. Torque</span>, property.value]
            case "torque_max":
                return [<span><Icon icon={"mdi--refresh"} />Max. Torque</span>, property.value]
            case "material":
                return [<span><Icon icon={"mdi--test-tube"} />Material</span>, property.value]
            case "color":
                return [<span><Icon icon={"mdi--color"} />Color</span>, property.type === "string" ? formatColor(property.value) : property.value.toString()]
            default:
                switch (property.type) {
                    case "string":
                        return [<><Icon icon="mdi--format-text" />{property.name}</>, formatValue(property.value)];
                    case "datetime":
                        return [<><Icon icon="mdi--date-range" />{property.name}</>, property.value.toString()];
                    case "float":
                        return [<><Icon icon="mdi--hashtag" />{property.name}</>, property.value];
                }
        }
    })();
    if (property.type === "float") {
        return <div className="flex flex-row gap-1"><div className="flex-none font-semibold">{first}</div>:<div>{second} {property.unit}</div></div>;
    } else {
        return <div className="flex flex-row gap-1"><div className="flex-none font-semibold">{first}</div>:<div>{second}</div></div>;
    }
};

export default PropertyComponent;