import { useCallback, useContext, useEffect, useState } from 'react';
import { Headline, Icon, ImageComponent } from '../../components/shared';
import { CartContext } from '../../context/cart';
import { Thing } from '../../api/resources';
import { getThing } from '../../api/things';
import { AxiosContext } from '../../context/axios';
import { UserNameAndUserId } from '../../components/shared/user';
import { RedButton } from '../../components/shared/button';

export const ThingInfoRow = ({ thing }: { thing: Thing }) => {
  const firstImage = thing.images[0];
  const firstImageContent = firstImage ? (
    <ImageComponent
      image={firstImage}
      defaultWidth={128}
      className="object-contain h-full w-full"
    />
  ) : (
    <span>
      <Icon icon="mdi--image-off-outline" />
    </span>
  );

  const { removeFromCart } = useContext(CartContext);

  const onCartClick = useCallback(() => {
    removeFromCart(thing.id);
  }, [removeFromCart, thing.id]);

  return (
    <div className="flex flex-row gap-4 flex-start items-center border border-secondary rounded-md p-1">
      <div className="flex w-16 h-16 items-center justify-center bg-brand-900 p-2 rounded-md">
        <a href={`/things/${thing.id}`}>{firstImageContent}</a>
      </div>
      <div className="w-32">
        <a href={`/things/${thing.id}`}>
          <h2 className="text-display text-xl mb-2">{thing.name}</h2>
        </a>
      </div>
      <div className="ml-auto"></div>
      <RedButton onClick={onCartClick}>
        <Icon className="" icon={'mdi--cart-remove'} />
      </RedButton>
    </div>
  );
};

export const ThingContainer = ({ thingId }: { thingId: string }) => {
  const [thing, setThing] = useState<Thing | null>(null);
  const axiosInstance = useContext(AxiosContext);

  useEffect(() => {
    if (!axiosInstance) {
      return;
    }
    getThing(axiosInstance, thingId).then((v) => setThing(v));
  }, [axiosInstance, thingId]);

  if (!thing) {
    return null;
  }
  return <ThingInfoRow thing={thing} />;
};

export const ShowCart = () => {
  const { cartByUser } = useContext(CartContext);

  return (
    <div>
      <Headline type="h2">Cart</Headline>
      {(cartByUser === null || Object.keys(cartByUser).length === 0) && (
        <p className="text-display">No entries yet</p>
      )}
      {cartByUser &&
        Object.entries(cartByUser).map(([ownerId, entries]) => {
          return (
            <div className="mb-4">
              <div className="flex flex-row gap-2 items-center mb-4">
                <Headline type="h3">Owned by</Headline>
                <a href={`/users/${ownerId}`}>
                  <UserNameAndUserId
                    userId={ownerId}
                    textColor="text-display"
                    imageBorderColor="border-display"
                  />
                </a>
              </div>
              <div className="flex flex-col gap-2">
                {entries.map((entry) => {
                  return <ThingContainer thingId={entry.thingId} key={entry.thingId} />;
                })}
              </div>
              <hr className="h-2 my-8 bg-primary border-2" />
            </div>
          );
        })}
    </div>
  );
};
