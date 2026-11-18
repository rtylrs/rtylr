import React, { useMemo } from 'react';
import { AvatarResolverProvider, type AvatarResolver } from '@voxire/ui';
import { AvatarPicture } from './AvatarPicture';
import { useAvatarDirectory } from './useAvatars';

/**
 * Makes every avatar in the portal that knows whose it is (a `seed` user id)
 * show that person's chosen avatar. Mounted once at the root.
 */
export const AvatarDirectoryProvider: React.FC<React.PropsWithChildren> = ({ children }) => {
  const directory = useAvatarDirectory();
  const resolver = useMemo<AvatarResolver>(
    () => ({
      render: (seed, size) => {
        const avatar = directory.get(seed);
        if (!avatar || avatar.kind === 'initials') return null;
        return <AvatarPicture avatar={avatar} size={size} />;
      },
      colorIndex: (seed) => {
        const avatar = directory.get(seed);
        return avatar?.kind === 'initials' ? avatar.color : undefined;
      },
    }),
    [directory]
  );
  return <AvatarResolverProvider resolver={resolver}>{children}</AvatarResolverProvider>;
};

export default AvatarDirectoryProvider;
