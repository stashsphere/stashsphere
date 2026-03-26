import { cssColorLuminance, HEX_RE, RGB_RE, HSL_RE } from './colors';

/**
 * Validates if a value is a valid CSS color and returns it as a string,
 * or undefined if invalid.
 *
 * Supports:
 * - Named CSS colors (e.g., "red", "blue", "aliceblue")
 * - Hex colors: #RGB, #RGBA, #RRGGBB, #RRGGBBAA
 * - RGB/RGBA: rgb(255, 0, 0), rgba(255, 0, 0, 0.5)
 * - HSL/HSLA: hsl(120, 100%, 50%), hsla(120, 100%, 50%, 0.5)
 */
export const validateColor = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined;
  }

  const trimmed = value.trim().toLowerCase();

  if (!trimmed) {
    return undefined;
  }

  // Check for named CSS colors
  if (trimmed in cssColorLuminance) {
    return trimmed;
  }

  // Check for hex colors
  if (HEX_RE.test(trimmed)) {
    return trimmed;
  }

  // Check for rgb/rgba
  const rgbMatch = trimmed.match(RGB_RE);
  if (rgbMatch) {
    const r = parseInt(rgbMatch[1]);
    const g = parseInt(rgbMatch[2]);
    const b = parseInt(rgbMatch[3]);
    // Validate RGB values are in range 0-255
    if (r >= 0 && r <= 255 && g >= 0 && g <= 255 && b >= 0 && b <= 255) {
      return trimmed;
    }
    return undefined;
  }

  // Check for hsl/hsla
  const hslMatch = trimmed.match(HSL_RE);
  if (hslMatch) {
    const h = parseInt(hslMatch[1]);
    const s = parseInt(hslMatch[2]);
    const l = parseInt(hslMatch[3]);
    // Validate HSL values: h: 0-360, s: 0-100, l: 0-100
    if (h >= 0 && h <= 360 && s >= 0 && s <= 100 && l >= 0 && l <= 100) {
      return trimmed;
    }
    return undefined;
  }

  return undefined;
};
