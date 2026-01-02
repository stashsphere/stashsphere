import { Thing } from '../../api/resources';
import { Icon } from '.';
import { PropertyList } from '../thing_details';
import { ImageComponent } from '../shared';
import { UserNameAndUserId } from './user';
import CartButton from './cart_button';

type ThingInfoProps = {
  thing: Thing;
  hideCart?: boolean;
};

export const ThingInfo = ({ thing, hideCart }: ThingInfoProps) => {
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
      <a href={`/things/${thing.id}`}>
        <div className="flex w-80 h-80 items-center justify-center bg-brand-900 p-2 rounded-md">
          {firstImageContent}
        </div>
      </a>
      <div className="w-80">
        <a href={`/things/${thing.id}`}>
          <h2 className="text-display text-xl mb-2">{thing.name}</h2>
        </a>
        <div className="flex flex-row gap-2 items-center justify-between">
          <UserNameAndUserId
            userId={thing.owner.id}
            textColor="text-display"
            imageBorderColor="border-display"
          />
          <h2 className="text-display">
            <Icon icon="mdi--animation" /> {thing.quantity} {thing.quantityUnit}
          </h2>
          {!hideCart && <CartButton thingId={thing.id} />}
        </div>
        <PropertyList properties={thing.properties} collapsable={true} keyWidth="8rem" />
      </div>
    </div>
  );
};
