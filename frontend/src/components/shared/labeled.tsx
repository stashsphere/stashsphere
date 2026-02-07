import { ReactNode } from 'react';

type LabeledProps = {
  label: string;
  children: ReactNode;
};

const isEmpty = (children: ReactNode): boolean => {
  if (children === null || children === undefined) return true;
  if (typeof children === 'string' && children.trim() === '') return true;
  return false;
};

export const Labeled = ({ label, children }: LabeledProps) => {
  return (
    <div className="border shadow-xs flex flex-col p-1 border-secondary rounded-sm">
      <div className="">
        <span className="uppercase font-semibold bg-primary p-1 rounded-sm text-xs text-onprimary">
          {label}
        </span>
      </div>
      {isEmpty(children) ? <span className="italic text-secondary">not provided</span> : children}
    </div>
  );
};
