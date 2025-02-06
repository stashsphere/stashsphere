import React from "react";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> { }

export const Button = ({ children }: ButtonProps) => {
  return (
    <button className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 focus:outline-none focus:ring focus:border-blue-300">
      {children}
    </button>
  );
};

export const PrimaryButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = "";
  if (disabled) {
    disabledClasses = "bg-gray-300 cursor-not-allowed text-gray-500";
  } else {
    disabledClasses = "bg-primary hover:bg-primary-hover focus:ring focus:border-primary-hover text-onprimary";
  }
  return (
    <button className={"rounded py-2 px-4".concat(" ", disabledClasses, " ", className || "")} {...rest}>
      {children}
    </button>
  )
}

export const SecondaryButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = "";
  if (disabled) {
    disabledClasses = "bg-gray-300 cursor-not-allowed text-gray-500";
  } else {
    disabledClasses = "bg-secondary hover:bg-secondary-hover focus:ring focus:border-secondary-hover text-onsecondary";
  }
  return (
    <button className={"rounded py-2 px-4".concat(" ",  disabledClasses, " ", className || "")} {...rest}>
      {children}
    </button>
  )
}

export const AccentButton = ({ children, className, disabled, ...rest }: ButtonProps) => {
  let disabledClasses = "";
  if (disabled) {
    disabledClasses = "bg-gray-300 cursor-not-allowed text-gray-500";
  } else {
    disabledClasses = "bg-accent hover:bg-accent-hover focus:ring focus:border-accent-hover text-onaccent";
  }
  return (
    <button className={"rounded py-2 px-4".concat(" ",  disabledClasses, " ", className || "")} {...rest}>
      {children}
    </button>
  )
}

export const NeutralButton = ({ children, className, ...rest }: ButtonProps) => {
  return (
    <button className={"bg-neutral text-onneutral hover:bg-neutral-hover rounded py-2 px-4 focus:ring focus:border-neutral-hover".concat(" ", className || "")} {...rest}>
      {children}
    </button>
  )
}

export const DangerButton = ({ children, className, ...rest }: ButtonProps) => {
  return (
    <button className={"bg-danger text-ondanger hover:bg-danger-hover rounded py-2 px-4 focus:ring focus:border-danger-hover".concat(" ", className || "")} {...rest}>
      {children}
    </button>
  )
}

// Color buttons

export const BlueButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-blue-500 text-white py-2 px-4 rounded hover:bg-blue-600 focus:outline-none focus:ring focus:border-blue-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const YellowButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-yellow-500 text-black py-2 px-4 rounded hover:bg-yellow-600 focus:outline-none focus:ring focus:border-yellow-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const GrayButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-gray-300 text-black py-2 px-4 rounded hover:bg-gray-600 focus:outline-none focus:ring focus:border-gray-300"
      {...rest}
    >
      {children}
    </button>
  );
};

export const RedButton = ({ children, ...rest }: ButtonProps) => {
  return (
    <button
      className="bg-red-500 text-white py-2 px-4 rounded hover:bg-red-600 focus:outline-none focus:ring focus:border-red-300"
      {...rest}
    >
      {children}
    </button>
  );
};
