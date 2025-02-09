import { useContext, useEffect, useState } from "react";
import { AxiosContext } from "../../context/axios";
import { PagedThings } from "../../api/resources";
import { getThings } from "../../api/things";
import { Pages } from "../../components/pages";
import ThingInfo from "../../components/thing_info";
import { PrimaryButton } from "../../components/button";

export const Things = () => {
  const axiosInstance = useContext(AxiosContext);
  const [things, setThings] = useState<PagedThings | undefined>(undefined);

  const [currentPage, setCurrentPage] = useState(0);

  useEffect(() => {
    if (axiosInstance === null) {
      return;
    }
    getThings(axiosInstance, currentPage)
      .then(setThings)
      .catch((reason) => {
        console.log(reason);
      });
  }, [axiosInstance, currentPage]);

  if (!things) {
    return <p>Loading...</p>;
  }

  return (
    <>
      <div className="flex flex-row flex-row-reverse">
        <a href="/things/create">
          <PrimaryButton>Add Thing</PrimaryButton>
        </a>
      </div>
      <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
        {things.things.map((thing) => (
          <ThingInfo thing={thing} key={thing.id} />
        ))}
      </div >
      {things.things.length > 0 &&
        <Pages
          currentPage={currentPage}
          onPageChange={(n) => setCurrentPage(n)}
          pages={things.totalPageCount}
        />
      }
    </>
  );
};
