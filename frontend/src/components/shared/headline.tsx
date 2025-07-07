interface HeadlineProps extends React.PropsWithChildren {
  type: 'h1' | 'h2' | 'h3';
}

export const Headline = ({ children, type }: HeadlineProps) => {
  switch (type) {
    case 'h1':
      return <h1 className="text-2xl font-bold text-accent">{children}</h1>;
    case 'h2':
      return <h2 className="text-xl font-semibold text-primary">{children}</h2>;
    case 'h3':
      return <h3 className="text-lg font-medium text-secondary">{children}</h3>;
    default:
      return <span>{children}</span>;
  }
};
