import { Icon } from './icon';

interface Props {
  quantity: number;
  unit: string;
  onChange?: (quantity: number, unit: string) => void;
}

const QuantityEditor = ({ quantity, unit, onChange }: Props) => {
  const onQuantityChange = (value: number) => {
    if (value < 0) {
      return;
    }
    if (onChange) {
      onChange(value, unit);
    }
  };

  const onUnitChange = (value: string) => {
    if (onChange) {
      onChange(quantity, value);
    }
  };

  const hideArrows =
    '[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none';

  const buttonClasses = 'flex items-center justify-center w-6 h-7 bg-primary text-onprimary';

  return (
    <div className="flex flex-col">
      <div className="flex flex-row text-display mb-4">
        <button
          className={`${buttonClasses} rounded-l-full`}
          onClick={() => onQuantityChange(quantity - 1)}
        >
          <Icon icon="mdi--minus" />
        </button>
        <input
          className={`w-16 ${hideArrows} text-right px-2`}
          type="number"
          min={0}
          step={1}
          onChange={(e) => onQuantityChange(Number(e.target.value))}
          value={quantity}
        />
        <button
          className={`${buttonClasses} rounded-r-full`}
          onClick={() => onQuantityChange(quantity + 1)}
        >
          <Icon icon="mdi--plus" />
        </button>
      </div>

      <div className="">
        <label htmlFor="name" className="block text-primary text-sm font-medium">
          Unit
        </label>
        <input
          type="text"
          id="name"
          name="name"
          value={unit}
          onChange={(e) => onUnitChange(e.target.value)}
          className="w-32 mt-1 p-2 text-display border border-gray-300 rounded-sm"
        />
      </div>
    </div>
  );
};
export default QuantityEditor;
