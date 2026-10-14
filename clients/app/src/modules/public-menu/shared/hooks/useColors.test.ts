import { describe, it, expect } from 'vitest';
import useColors from './useColors';

describe('useColors hook', () => {
  it('returns color palette object with expected keys', () => {
    const colors = useColors();
    expect(colors).toEqual({
      primary: '#040428',
      secondary: '#4485c4',
      accent: '#03AED2',
      darkColor: '#016B63',
      background: '#ffffff',
      bgColor: '#ffffff',
      light: '#fff',
      dark: '#000',
      itemBackground: '#373737',
      textColor: '#fff',
      subTextColor: '#B0B0B0',
      lightGray: '#D9D9D9',
    });
  });
});
