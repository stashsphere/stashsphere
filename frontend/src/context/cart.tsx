import React from 'react';
import { Cart } from '../api/resources';
import { CartByUser } from '../hooks/useCart';

export type CartContextType = {
  cart: Cart | null;
  addToCart: (thingId: string) => void;
  removeFromCart: (thingId: string) => void;
  clearCart: () => void;
  cartByUser: CartByUser | null;
};

export const CartContext = React.createContext<CartContextType>({} as CartContextType);
