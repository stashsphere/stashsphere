import { Icon } from '../shared';

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
    <div className="flex flex-row items-end gap-4">
      <div className="flex flex-col">
        <label className="block text-primary text-sm font-medium mb-1">Quantity</label>
        <div className="flex flex-row text-display">
          <button
            className={`${buttonClasses} rounded-l-full`}
            onClick={() => onQuantityChange(quantity - 1)}
          >
            <Icon icon="mdi--minus" />
          </button>
          <input
            className={`w-16 ${hideArrows} text-right px-2 border-y border-gray-300`}
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
      </div>

      <div className="flex flex-col">
        <label htmlFor="unit" className="block text-primary text-sm font-medium mb-1">
          Unit
        </label>
        <input
          type="text"
          id="unit"
          name="unit"
          value={unit}
          onChange={(e) => onUnitChange(e.target.value)}
          className="w-32 p-2 text-display border border-gray-300 rounded-sm"
          placeholder="e.g. pieces, kg, liters"
        />
      </div>
    </div>
  );
};
export default QuantityEditor;
