import { Thing } from '../api/resources';
import { Icon } from './shared';
import PropertyList from './thing_details/property_list';
import { ImageComponent } from './image';

type ThingInfoProps = {
  thing: Thing;
};

const ThingInfo = ({ thing }: ThingInfoProps) => {
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
        <a href={`/things/${thing.id}`}>
          <h2 className="text-display text-xl mb-2">{thing.name}</h2>
          <div className="flex flex-row gap-2">
            <h2 className="text-display">
              <Icon icon="mdi--user" /> {thing.owner.name}
            </h2>
            <h2 className="text-display">
              <Icon icon="mdi--animation" /> {thing.quantity} {thing.quantityUnit}
            </h2>
          </div>
        </a>
        <PropertyList properties={thing.properties} collapsable={true} keyWidth="8rem" />
      </div>
    </div>
  );
};

export default ThingInfo;
