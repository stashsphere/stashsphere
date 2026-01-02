import { ReactNode } from 'react';

export const Toggle = ({
  value,
  onChange,
  children,
}: {
  value: boolean;
  onChange: (newValue: boolean) => void;
  children: ReactNode;
}) => {
  return (
    <label className="inline-flex items-center cursor-pointer">
      <input
        type="checkbox"
        checked={value}
        className="sr-only peer"
        onChange={(e) => onChange(e.target.checked)}
      />
      <div className="relative w-11 h-6 bg-secondary-500 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-primary after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-primary after:border-primary-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-secondary-600"></div>
      {children}
    </label>
  );
};
