import React from 'react';

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement>;

export const Button = ({ children }: ButtonProps) => {
  return (
    <button className="bg-blue-500 text-white py-2 px-4 rounded-sm hover:bg-blue-600 focus:outline-hidden focus:ring-3 focus:border-blue-300">
      {children}
    </button>
  );
};

export const PrimaryButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = '';
  if (disabled) {
    disabledClasses = 'bg-gray-300 cursor-not-allowed text-gray-500';
  } else {
    disabledClasses =
      'bg-primary hover:bg-primary-hover focus:ring-3 focus:border-primary-hover text-onprimary';
  }
  return (
    <button
      className={'rounded-sm py-2 px-4'.concat(' ', disabledClasses, ' ', className || '')}
      {...rest}
    >
      {children}
    </button>
  );
};

export const SecondaryButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = '';
  if (disabled) {
    disabledClasses = 'bg-gray-300 cursor-not-allowed text-gray-500';
  } else {
    disabledClasses =
      'bg-secondary hover:bg-secondary-hover focus:ring-3 focus:border-secondary-hover text-onsecondary';
  }
  return (
    <button
      className={'rounded-sm py-2 px-4'.concat(' ', disabledClasses, ' ', className || '')}
      {...rest}
    >
      {children}
    </button>
  );
};

export const AccentButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = '';
  if (disabled) {
    disabledClasses = 'bg-gray-300 cursor-not-allowed text-gray-500';
  } else {
    disabledClasses =
      'bg-accent hover:bg-accent-hover focus:ring-3 focus:border-accent-hover text-onaccent';
  }
  return (
    <button
      className={'rounded-sm py-2 px-4'.concat(' ', disabledClasses, ' ', className || '')}
      {...rest}
    >
      {children}
    </button>
  );
};

export const NeutralButton = ({ children, className, ...rest }: ButtonProps) => {
  return (
    <button
      className={'bg-neutral text-onneutral hover:bg-neutral-hover rounded-sm py-2 px-4 focus:ring-3 focus:border-neutral-hover'.concat(
        ' ',
        className || ''
      )}
      {...rest}
    >
      {children}
    </button>
  );
};

export const DangerButton = ({ children, className, ...rest }: ButtonProps) => {
  return (
    <button
      className={'bg-danger text-ondanger hover:bg-danger-hover rounded-sm py-2 px-4 focus:ring-3 focus:border-danger-hover'.concat(
        ' ',
        className || ''
      )}
      {...rest}
    >
      {children}
    </button>
  );
};

// Color buttons

export const BlueButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-blue-500 text-white py-2 px-4 rounded-sm hover:bg-blue-600 focus:outline-hidden focus:ring-3 focus:border-blue-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const YellowButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-yellow-500 text-black py-2 px-4 rounded-sm hover:bg-yellow-600 focus:outline-hidden focus:ring-3 focus:border-yellow-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const GrayButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-gray-300 text-black py-2 px-4 rounded-sm hover:bg-gray-600 focus:outline-hidden focus:ring-3 focus:border-gray-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const RedButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-red-500 text-white py-2 px-4 rounded-sm hover:bg-red-600 focus:outline-hidden focus:ring-3 focus:border-red-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const GreenButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-green-600 text-white py-2 px-4 rounded-sm hover:bg-green-500 focus:outline-hidden focus:ring-3 focus:border-green-300"
      {...rest}
    >
      {children}
    </button>
  );
};
