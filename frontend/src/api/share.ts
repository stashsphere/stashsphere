import { Axios } from "axios";
import { Share } from "./resources";

export interface CreateShareParams {
    targetUserId: string;
    objectId: string;
}

export const shareObject = async (axios: Axios, params: CreateShareParams) => {
    const response = await axios.post("/shares", params, { headers: {
        "Content-Type": "application/json"
    }});

    const share = response.data as Share;
    return share;
}

export const deleteShare = async (axios: Axios, shareId: string) => {
    await axios.delete(`/shares/${shareId}`)
};