// collection of known units with proper utf-8 rendering
export const UNIT_LABELS: Record<string, string> = {
  celsius: '°C',
  fahrenheit: '°F',
  kelvin: 'K',
  ohm: 'Ω',
};

export const PREFIX_LABELS: Record<string, string> = {
  micro: 'μ',
};

export const substituteUnits = (unit: string): string => {
  let res = unit;
  for (const [key, substitute] of Object.entries(UNIT_LABELS)) {
    if (res.toLowerCase().includes(key)) {
      res = res.replace(new RegExp(key, 'gi'), substitute);
    }
  }

  for (const [key, substitute] of Object.entries(PREFIX_LABELS)) {
    if (res.includes(key)) {
      res = res.replace(key, substitute);
    }
  }

  return res;
};
