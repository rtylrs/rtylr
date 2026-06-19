import { cdnImageUrl } from '@voxire/config';

export const ASSETS = {
  logos: {
    dark: cdnImageUrl('https://cdn.voxire.dev/web_content/rtylr/logo-dark.webp') ?? '',
    light: cdnImageUrl('https://cdn.voxire.dev/web_content/rtylr/logo-light.webp') ?? '',
  },
};
