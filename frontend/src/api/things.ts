import { Axios } from "axios";
import { PagedThings, Thing } from "./resources";

export const getThings = async (axios: Axios, currentPage: number) => {
    const response = await axios.get(`/things?page=${currentPage}`, { headers: {
            "Content-Type": "application/json"
    }});
    
    if (response.status != 200) {
        throw `Got error ${response}`
    }
   
    const things = response.data as PagedThings;
    return things;
}

export const getThing = async (axios: Axios, id: string) => {
    const response = await axios.get(`/things/${id}`, { headers: {
            "Content-Type": "application/json"
    }});
    
    if (response.status != 200) {
        throw `Got error ${response}`
    }
    return response.data as Thing
}

export type CreatePropertyStringParam = {
    name: string;
    value: string;
    type: "string";
}
export type CreatePropertyFloatParam = { name: string; value: number; type: "float"; unit?: string; }
export type CreatePropertyDatetimeParam = { name: string; value: string; type: "datetime"; }
export type CreatePropertyParam  = CreatePropertyDatetimeParam | CreatePropertyStringParam | CreatePropertyFloatParam;

export interface CreateThingParams {
    name: string,
    privateNote: string;
    description: string;
    imagesIds: string[];
    properties: CreatePropertyParam[];
    quantity: number;
    quantityUnit: string;
}

export const createThing = async (axios: Axios, params: CreateThingParams): Promise<Thing> => {
    const response = await axios.post("/things", params, { headers: {
        "Content-Type": "application/json"
    }});
    
    const thing = response.data as Thing;
    return thing;
}

export interface UpdateThingParams extends CreateThingParams {}


export const updateThing = async (axios: Axios, id: string, params: UpdateThingParams) => {
    const response = await axios.patch("/things/" + id, params, { headers: {
        "Content-Type": "application/json"
    }});
    
    const thing = response.data as Thing;
    return thing;
}