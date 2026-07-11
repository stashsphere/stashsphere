import { useContext } from 'react';
import { PublicThing } from '../../api/resources';
import { ThingImages } from '../thing_details/thing_images';
import { PropertyList } from '../shared/property_list';
import { Headline } from '../shared';
import { PropertySchemaCollectionContext } from '../../context/property_schema';

interface PublicThingDetailsProps {
  thing: PublicThing;
}

export const PublicThingDetails = ({ thing }: PublicThingDetailsProps) => {
  const schemaCollection = useContext(PropertySchemaCollectionContext);

  return (
    <div className="flex flex-col gap-8">
      <div className="flex flex-row justify-between">
        <Headline type="h1">{thing.name}</Headline>
      </div>

      <div className="flex flex-col md:flex-row gap-6">
        <div className="flex-1">
          <ThingImages images={thing.images} />
        </div>
        <div className="flex flex-col flex-1 gap-6">
          <div>
            <Headline type="h2">Owner</Headline>
            <p className="text-display">{thing.ownerName}</p>
          </div>
          <div>
            <Headline type="h2">Quantity</Headline>
            <p className="text-display text-l">
              {thing.quantity} {thing.quantityUnit}
            </p>
          </div>
          <PropertyList properties={thing.properties} schemaCollection={schemaCollection} />
          <div>
            <Headline type="h2">Description</Headline>
            <div className="text-display">{thing.description}</div>
          </div>
        </div>
      </div>
    </div>
  );
};
