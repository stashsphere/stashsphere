import { useCallback, useEffect, useMemo, useState } from 'react';
import { Cart, CartEntry } from '../api/resources';
import { getCart, putCart } from '../api/cart';
import { Axios } from 'axios';

export type CartByUser = Record<string, CartEntry[]>;

export const useCart = (
  axios: Axios | null,
  loggedIn: boolean
): [
  Cart | null,
  (thingId: string) => void,
  (thingId: string) => void,
  () => void,
  CartByUser | null,
] => {
  const [cart, setCart] = useState<Cart | null>(null);
  useEffect(() => {
    if (axios === null) {
      return;
    }
    if (!loggedIn) {
      setCart(null);
      return;
    }
    getCart(axios).then((cart) => setCart(cart));
  }, [axios, loggedIn]);

  const addThing = useCallback(
    (thingId: string) => {
      if (axios === null) {
        return;
      }
      if (cart === null) {
        return;
      }
      putCart(axios, [thingId, ...cart.entries.map((v) => v.thingId)]).then(setCart);
    },
    [axios, cart]
  );

  const removeThing = useCallback(
    (thingId: string) => {
      if (axios === null) {
        return;
      }
      if (cart === null) {
        return;
      }
      putCart(axios, [...cart.entries.map((v) => v.thingId).filter((v) => v !== thingId)]).then(
        setCart
      );
    },
    [axios, cart]
  );

  const clearCart = useCallback(() => {
    if (axios === null) {
      return;
    }
    putCart(axios, []).then(setCart);
  }, [axios]);

  const cartByUser = useMemo(() => {
    const res: CartByUser = {};
    if (!cart) {
      return null;
    }
    for (const entry of cart.entries) {
      if (entry.ownerId in res) {
        res[entry.ownerId] = [...res[entry.ownerId], entry];
      } else {
        res[entry.ownerId] = [entry];
      }
    }
    return res;
  }, [cart]);

  return [cart, addThing, removeThing, clearCart, cartByUser];
};
