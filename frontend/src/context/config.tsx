import React from "react";
import { Config } from "../config/config";

export const ConfigContext = React.createContext<Config>({
    apiHost: "http://127.0.0.1"
});