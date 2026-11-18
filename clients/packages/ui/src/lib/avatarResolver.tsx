import React from "react";

/**
 * Looks up the avatar a person chose, by their id (`seed`).
 *
 * The app provides it once at the root, so every avatar that knows whose it
 * is shows the person's own choice without each screen fetching anything.
 * Packages stay free of the API: they only ask.
 */
export interface AvatarResolver {
  /** A picture (photo, illustration, emoji) to draw instead of initials, or null. */
  render: (seed: string, sizePx: number) => React.ReactNode | null;
  /** The palette colour the person picked for their initials, if they did. */
  colorIndex: (seed: string) => number | undefined;
}

const AvatarResolverContext = React.createContext<AvatarResolver | null>(null);

export const AvatarResolverProvider: React.FC<
  React.PropsWithChildren<{ resolver: AvatarResolver }>
> = ({ resolver, children }) => (
  <AvatarResolverContext.Provider value={resolver}>
    {children}
  </AvatarResolverContext.Provider>
);

export function useAvatarResolver(): AvatarResolver | null {
  return React.useContext(AvatarResolverContext);
}
