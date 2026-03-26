import { cssColorLuminance, HEX_RE } from './colors';

const linearize = (c: number): number => {
  const s = c / 255;
  return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
};

const rgbLuminance = (r: number, g: number, b: number): number => {
  return 0.2126 * linearize(r) + 0.7152 * linearize(g) + 0.0722 * linearize(b);
};

const DEFAULT_BG = '#808080';

export const luminanceToBackground = (l: number): string => {
  if (l < 0.3) return '#ffffff';
  if (l < 0.5) return '#d0d0d0';
  if (l < 0.6) return '#303030';
  return '#000000';
};

export const contrastBackground = (color: string | undefined): string => {
  if (!color) {
    return DEFAULT_BG;
  }

  const trimmed = color.trim().toLowerCase();

  if (trimmed in cssColorLuminance) {
    return luminanceToBackground(cssColorLuminance[trimmed]);
  }

  const m = trimmed.match(HEX_RE);
  if (m) {
    const r = parseInt(m[1], 16);
    const g = parseInt(m[2], 16);
    const b = parseInt(m[3], 16);
    return luminanceToBackground(rgbLuminance(r, g, b));
  }

  return DEFAULT_BG;
};
