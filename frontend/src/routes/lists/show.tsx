import { useParams } from "react-router-dom";
import { ListDetails } from "../../components/list_details";

export const ShowList = () => {
    const { listId } = useParams();
    if (listId === undefined) {
        return <p>invalid id</p>
    } else {
        return (
            <ListDetails id={listId}/>
        )
    }
}