import React, { useEffect, useState } from 'react';
import { avatarColorAt } from '@voxire/ui';
import type { ChosenAvatar } from './avatarModel';
import { cachedIllustration, renderIllustration } from './illustrated';

/**
 * The picture for a chosen avatar that is not plain initials: a photo, an
 * illustration or an emoji. Fills its parent circle.
 */
export const AvatarPicture: React.FC<{ avatar: ChosenAvatar; size: number }> = ({
  avatar,
  size,
}) => {
  if (avatar.kind === 'photo' && avatar.url) {
    return <PhotoPicture url={avatar.url} />;
  }
  if (avatar.kind === 'illustrated' && avatar.style && avatar.seed) {
    return <Illustration style={avatar.style} seed={avatar.seed} />;
  }
  if (avatar.kind === 'emoji' && avatar.emoji) {
    const color = avatarColorAt(avatar.color ?? 0);
    return (
      <span
        aria-hidden="true"
        className="flex h-full w-full items-center justify-center leading-none"
        style={{ background: color.background, fontSize: Math.round(size * 0.58) }}
      >
        {avatar.emoji}
      </span>
    );
  }
  return null;
};

const PhotoPicture: React.FC<{ url: string }> = ({ url }) => {
  const [failed, setFailed] = useState(false);
  if (failed)
    return <span className="h-full w-full" style={{ background: 'var(--surface-2, #e5e7eb)' }} />;
  return (
    <img
      src={url}
      alt=""
      loading="lazy"
      decoding="async"
      draggable={false}
      onError={() => setFailed(true)}
      className="h-full w-full object-cover"
    />
  );
};

const Illustration: React.FC<{ style: string; seed: string }> = ({ style, seed }) => {
  const [uri, setUri] = useState(() => cachedIllustration(style, seed));
  useEffect(() => {
    let live = true;
    const cached = cachedIllustration(style, seed);
    if (cached) {
      setUri(cached);
      return undefined;
    }
    setUri(undefined);
    renderIllustration(style, seed)
      .then((next) => {
        if (live) setUri(next);
      })
      .catch(() => undefined);
    return () => {
      live = false;
    };
  }, [style, seed]);
  if (!uri)
    return <span className="h-full w-full" style={{ background: 'var(--surface-2, #e5e7eb)' }} />;
  return <img src={uri} alt="" draggable={false} className="h-full w-full object-cover" />;
};

export default AvatarPicture;
