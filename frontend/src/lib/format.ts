export const formatHumanUnit = (value: number): string => {
  const formatter = Intl.NumberFormat('en', { notation: 'compact' });
  return formatter.format(value);
};
