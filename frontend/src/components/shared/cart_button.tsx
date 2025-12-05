import { useCallback, useContext, useMemo } from 'react';
import { GreenButton, PrimaryButton } from './button';
import { Icon } from './icon';
import { CartContext } from '../../context/cart';

const CartButton = ({ thingId }: { thingId: string }) => {
  const { cart, addToCart, removeFromCart } = useContext(CartContext);

  const isPartOfCart = useMemo(() => {
    if (!cart) {
      return false;
    }
    return cart.entries.map((e) => e.thingId).includes(thingId);
  }, [cart, thingId]);

  const onCartClick = useCallback(() => {
    if (isPartOfCart) {
      removeFromCart(thingId);
    } else {
      addToCart(thingId);
    }
  }, [isPartOfCart, removeFromCart, thingId, addToCart]);

  return isPartOfCart ? (
    <GreenButton onClick={onCartClick}>
      <Icon className="" icon={'mdi--cart-remove'} />
    </GreenButton>
  ) : (
    <PrimaryButton onClick={onCartClick}>
      <Icon className="" icon={'mdi--cart-plus'} />
    </PrimaryButton>
  );
};

export default CartButton;
