import { Profile } from "../api/resources";
import { Icon } from "./icon";

type ProfileListProps = {
    profiles: Profile[];
    onClick?: (profile: Profile) => void;
    hintText?: string;
};


type ProfileListItemProps = {
    profile: Profile;
    hintText?: string;
};

const ProfileListItem = (props: ProfileListItemProps) => {
    return (
        <div className="flex items-center p-3 -mt-2 text-sm text-gray-600 transition-colors duration-300 transform dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 dark:hover:text-white">
            <div className="w-16 h-16">
                <Icon icon="mdi--image-off-outline" height={"100%"} width={"100%"}/>
            </div>
            <div className="mx-1">
                <h1 className="text-sm font-semibold text-gray-700 dark:text-gray-200">{props.profile.name}</h1>
                <p className="text-sm text-gray-500 dark:text-gray-400">{props.profile.email}</p>
            </div>
            {props.hintText}
        </div>
    )
}

export const ProfileList = (props: ProfileListProps) => {
    return (
        <ul className="border border-gray-300">
            {
                props.profiles.map((profile, index) => (
                    <li onClick={() => props.onClick && props.onClick(profile)} key={profile.id}>
                        <ProfileListItem key={index} profile={profile} hintText={props.hintText}/>
                    </li>
                ))
            }
        </ul>
    )
}