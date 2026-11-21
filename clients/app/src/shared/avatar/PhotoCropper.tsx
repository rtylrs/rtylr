import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../components/ui/button';

/** The circle the user frames their photo in, in CSS pixels. */
const VIEW = 224;
/** The stored photo is this many pixels square: sharp at every avatar size. */
export const AVATAR_PHOTO_PX = 256;

interface Frame {
  /** Zoom on top of "cover": 1 fills the circle exactly. */
  zoom: number;
  /** Offset of the image centre from the circle centre, in view pixels. */
  x: number;
  y: number;
}

/** Keeps the image covering the whole circle, so no empty corner is ever saved. */
function clampFrame(frame: Frame, natural: { w: number; h: number }): Frame {
  const cover = Math.max(VIEW / natural.w, VIEW / natural.h) * frame.zoom;
  const maxX = Math.max(0, (natural.w * cover - VIEW) / 2);
  const maxY = Math.max(0, (natural.h * cover - VIEW) / 2);
  return {
    zoom: frame.zoom,
    x: Math.min(maxX, Math.max(-maxX, frame.x)),
    y: Math.min(maxY, Math.max(-maxY, frame.y)),
  };
}

/** Draws the framed circle's square to a canvas at the stored size. */
export function cropToCanvas(
  image: CanvasImageSource,
  natural: { w: number; h: number },
  frame: Frame,
  out = AVATAR_PHOTO_PX
): HTMLCanvasElement {
  const cover = Math.max(VIEW / natural.w, VIEW / natural.h) * frame.zoom;
  const canvas = document.createElement('canvas');
  canvas.width = out;
  canvas.height = out;
  const ctx = canvas.getContext('2d');
  if (!ctx) return canvas;
  // Source square, in image pixels, under the circle.
  const side = VIEW / cover;
  const sx = natural.w / 2 - frame.x / cover - side / 2;
  const sy = natural.h / 2 - frame.y / cover - side / 2;
  ctx.imageSmoothingQuality = 'high';
  ctx.drawImage(image, sx, sy, side, side, 0, 0, out, out);
  return canvas;
}

interface PhotoCropperProps {
  file: File;
  /** Called with the cropped square, ready to upload. */
  onCropped: (blob: Blob) => void;
  onCancel: () => void;
  busy?: boolean;
}

/**
 * Frame a photo in a circle: drag to move, slide (or scroll, or arrow keys)
 * to zoom. Everything happens in the browser; only the 256px result uploads.
 */
export const PhotoCropper: React.FC<PhotoCropperProps> = ({ file, onCropped, onCancel, busy }) => {
  const { t } = useTranslation();
  const [src, setSrc] = useState<string>();
  const [natural, setNatural] = useState<{ w: number; h: number }>();
  const [frame, setFrame] = useState<Frame>({ zoom: 1, x: 0, y: 0 });
  const imageRef = useRef<HTMLImageElement>(null);
  const drag = useRef<{ id: number; x: number; y: number; start: Frame } | null>(null);

  useEffect(() => {
    const url = URL.createObjectURL(file);
    setSrc(url);
    setFrame({ zoom: 1, x: 0, y: 0 });
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const update = useCallback(
    (next: Frame) => setFrame(natural ? clampFrame(next, natural) : next),
    [natural]
  );

  const cover = natural ? Math.max(VIEW / natural.w, VIEW / natural.h) * frame.zoom : 1;

  const save = (): void => {
    if (!imageRef.current || !natural) return;
    const canvas = cropToCanvas(imageRef.current, natural, frame);
    canvas.toBlob(
      (blob) => {
        if (blob) onCropped(blob);
      },
      'image/webp',
      0.9
    );
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div
        role="img"
        aria-label={t('settings.avatar.photoFrame')}
        tabIndex={0}
        className="relative cursor-grab touch-none overflow-hidden rounded-full active:cursor-grabbing"
        style={{ width: VIEW, height: VIEW, background: 'var(--surface-2, #e5e7eb)' }}
        onPointerDown={(e) => {
          e.currentTarget.setPointerCapture(e.pointerId);
          drag.current = { id: e.pointerId, x: e.clientX, y: e.clientY, start: frame };
        }}
        onPointerMove={(e) => {
          const d = drag.current;
          if (!d || d.id !== e.pointerId) return;
          update({ ...d.start, x: d.start.x + e.clientX - d.x, y: d.start.y + e.clientY - d.y });
        }}
        onPointerUp={() => {
          drag.current = null;
        }}
        onWheel={(e) =>
          update({ ...frame, zoom: Math.min(3, Math.max(1, frame.zoom - e.deltaY / 500)) })
        }
        onKeyDown={(e) => {
          const step = 8;
          const moves: Record<string, Partial<Frame>> = {
            ArrowLeft: { x: frame.x + step },
            ArrowRight: { x: frame.x - step },
            ArrowUp: { y: frame.y + step },
            ArrowDown: { y: frame.y - step },
            '+': { zoom: Math.min(3, frame.zoom + 0.1) },
            '-': { zoom: Math.max(1, frame.zoom - 0.1) },
          };
          const move = moves[e.key];
          if (move) {
            e.preventDefault();
            update({ ...frame, ...move });
          }
        }}
      >
        {src && (
          <img
            ref={imageRef}
            src={src}
            alt=""
            draggable={false}
            onLoad={(e) =>
              setNatural({ w: e.currentTarget.naturalWidth, h: e.currentTarget.naturalHeight })
            }
            className="pointer-events-none absolute left-1/2 top-1/2 max-w-none select-none"
            style={
              natural
                ? {
                    width: natural.w * cover,
                    height: natural.h * cover,
                    transform: `translate(calc(-50% + ${frame.x}px), calc(-50% + ${frame.y}px))`,
                  }
                : { opacity: 0 }
            }
          />
        )}
      </div>
      <label className="flex w-full max-w-[260px] items-center gap-3 text-sm">
        <span>{t('settings.avatar.zoom')}</span>
        <input
          type="range"
          min={1}
          max={3}
          step={0.01}
          value={frame.zoom}
          onChange={(e) => update({ ...frame, zoom: Number(e.target.value) })}
          className="flex-1"
        />
      </label>
      <p className="text-xs text-[var(--dash-text-dim)]">{t('settings.avatar.dragHint')}</p>
      <div className="flex gap-2">
        <Button variant="outline" onClick={onCancel} disabled={busy}>
          {t('settings.avatar.choosePhotoAgain')}
        </Button>
        <Button onClick={save} disabled={busy || !natural}>
          {busy ? t('settings.avatar.uploading') : t('settings.avatar.usePhoto')}
        </Button>
      </div>
    </div>
  );
};

export default PhotoCropper;
