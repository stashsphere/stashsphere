import axios from "axios";
import { config as devConfig } from "./config.development.ts";

export interface Config {
    apiHost: string;
}

export const getConfig = async (): Promise<Config> => {
    if (process.env.NODE_ENV == "development") {
        return devConfig;
    } else {
        const resp = await axios.get("/config.json");
        return resp.data as Config;
    }
};
