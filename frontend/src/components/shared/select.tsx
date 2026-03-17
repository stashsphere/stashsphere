import React from 'react';

type SelectProps = React.SelectHTMLAttributes<HTMLSelectElement>;

export const Select = ({ children, className, ...rest }: SelectProps) => {
  return (
    <select
      className={'text-display border border-secondary shadow-xs focus:border-secondary rounded-sm px-2 py-1'.concat(
        ' ',
        className || ''
      )}
      {...rest}
    >
      {children}
    </select>
  );
};
