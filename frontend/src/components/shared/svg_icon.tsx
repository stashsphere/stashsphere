import { useEffect, useRef } from 'react';

type Props = {
  icon: string;
  className?: string;
  color?: string;
};

const SUPPORTED_SETS = ['mdi'];

const applyColor = (obj: HTMLObjectElement, color: string) => {
  const doc = obj.contentDocument;
  if (!doc) return;
  doc.querySelectorAll('[fill]').forEach((el) => {
    el.setAttribute('fill', color);
  });
  doc.querySelectorAll('[stroke]').forEach((el) => {
    if (el.getAttribute('stroke') !== 'none') {
      el.setAttribute('stroke', color);
    }
  });
};

export const SchemaIcon = ({ icon, className = '', color }: Props) => {
  const ref = useRef<HTMLObjectElement>(null);

  const colonIdx = icon.indexOf(':');
  const set = colonIdx === -1 ? '' : icon.slice(0, colonIdx);
  const name = colonIdx === -1 ? icon : icon.slice(colonIdx + 1);

  useEffect(() => {
    if (color && ref.current) {
      applyColor(ref.current, color);
    }
  }, [color]);

  if (!name || !set || !SUPPORTED_SETS.includes(set)) return null;

  const url = `/icons/${set}/${name}.svg`;

  return (
    <object
      ref={ref}
      data={url}
      type="image/svg+xml"
      onLoad={() => {
        if (color && ref.current) applyColor(ref.current, color);
      }}
      aria-hidden="true"
      className={`inline-block shrink-0 pointer-events-none ${className} w-5 h-5`}
    />
  );
};
