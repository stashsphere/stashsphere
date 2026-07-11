import { useContext } from 'react';
import { PublicList, PublicThing } from '../../api/resources';
import { Headline, Icon } from '../shared';
import { ImageComponent } from '../shared';
import { PropertyList } from '../shared/property_list';
import { PropertySchemaCollectionContext } from '../../context/property_schema';

interface PublicListDetailsProps {
  list: PublicList;
}

const PublicThingInfo = ({ thing }: { thing: PublicThing }) => {
  const schemaCollection = useContext(PropertySchemaCollectionContext);
  const firstImage = thing.images[0];
  const firstImageContent = firstImage ? (
    <ImageComponent
      image={firstImage}
      defaultWidth={512}
      className="object-contain h-full w-full"
    />
  ) : (
    <span>
      <Icon icon="mdi--image-off-outline" />
    </span>
  );

  return (
    <div className="flex flex-col gap-4 flex-start items-start border border-secondary rounded-md p-1">
      <div className="flex w-80 h-80 items-center justify-center bg-brand-900 p-2 rounded-md">
        {firstImageContent}
      </div>
      <div className="w-80">
        <h2 className="text-display text-xl mb-2">{thing.name}</h2>
        <div className="flex flex-row gap-2 items-center justify-between">
          <h2 className="text-display">
            <Icon icon="mdi--animation" /> {thing.quantity} {thing.quantityUnit}
          </h2>
        </div>
        <PropertyList
          properties={thing.properties}
          schemaCollection={schemaCollection}
          collapsable
        />
      </div>
    </div>
  );
};

export const PublicListDetails = ({ list }: PublicListDetailsProps) => {
  return (
    <>
      <div className="flex flex-row justify-between mb-4">
        <h1 className="text-2xl text-accent">{list.name}</h1>
      </div>
      <div>
        <Headline type="h2">Owner</Headline>
        <p className="text-display">{list.ownerName}</p>
      </div>
      <div className="flex flex-row gap-4 mt-4 flex-wrap justify-center">
        {list.things.map((thing) => (
          <PublicThingInfo thing={thing} key={thing.id} />
        ))}
      </div>
    </>
  );
};
