import { Axios } from "axios";
import { FriendRequest, FriendRequestResponse, FriendShips } from "./resources";

export const getFriendShips = async (axios: Axios) => {
    const response = await axios.get("/friends");

    const friendShips = response.data as FriendShips;
    return friendShips;
}

export const getFriendRequests = async (axios: Axios) => {
    const response = await axios.get(
        "/friend_requests"
    );

    const friendRequests = response.data as FriendRequestResponse;
    return friendRequests;
}

export const sendFriendRequest = async (axios: Axios, receiverId: string) => {
    const response = await axios.post("/friend_requests", {
        receiverId
    }, {
        headers: {
            "Content-Type": "application/json"
        }
    });

    const request = response.data as FriendRequest;
    return request;
}

export const reactToFriendRequest = async (axios: Axios, requestId: string, accept: boolean) => {
    const response = await axios.patch(`/friend_requests/${requestId}`, {
        accept
    }, {
        headers: {
            "Content-Type": "application/json"
        }
    });

    const request = response.data as FriendRequest;
    return request;
}

export const deleteFriendShip = async (axios: Axios, friendId: string) => {
    await axios.delete(`/friends/${friendId}`);
    return null;   
}