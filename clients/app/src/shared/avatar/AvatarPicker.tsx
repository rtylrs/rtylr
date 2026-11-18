import React, { useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Shuffle, Upload } from 'lucide-react';
import { uploadAsset } from '@voxire/api';
import {
  AVATAR_PALETTE,
  avatarColor,
  avatarColorAt,
  avatarColorIndex,
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@voxire/ui';
import {
  AVATAR_EMOJI,
  AVATAR_STYLE_GROUPS,
  AVATAR_STYLES,
  randomAvatarSeed,
  type ChosenAvatar,
} from './avatarModel';
import { AvatarPicture } from './AvatarPicture';
import { MOODS, MOODS_STYLE_ID } from './moods';
import { PhotoCropper } from './PhotoCropper';
import { useUpdateAvatar } from './useAvatars';

type Tab = 'photo' | 'illustrated' | 'initials' | 'emoji';
const TABS: Tab[] = ['photo', 'illustrated', 'initials', 'emoji'];
const MAX_PHOTO_BYTES = 5 * 1024 * 1024;
const PHOTO_TYPES = ['image/jpeg', 'image/png', 'image/webp'];
const GRID = 12;

const freshSeeds = (): string[] => Array.from({ length: GRID }, randomAvatarSeed);

export function initialsOf(name: string | undefined): string {
  const parts = (name ?? '').trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return '?';
  const first = parts[0][0] ?? '';
  const last = parts.length > 1 ? (parts[parts.length - 1][0] ?? '') : '';
  return (first + last).toUpperCase();
}

/** A chosen avatar (or the default) as a circle, for the preview and the page. */
export const AvatarPreview: React.FC<{
  avatar: ChosenAvatar | null | undefined;
  userId: string;
  name?: string;
  size: number;
}> = ({ avatar, userId, name, size }) => {
  const color =
    avatar?.kind === 'initials' && avatar.color !== undefined
      ? avatarColorAt(avatar.color)
      : avatarColor(userId);
  const picture = avatar && avatar.kind !== 'initials';
  return (
    <span
      className="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full font-semibold"
      style={{
        width: size,
        height: size,
        background: picture ? undefined : color.background,
        color: color.foreground,
        fontSize: Math.round(size * 0.38),
      }}
      data-avatar-preview={avatar?.kind ?? 'default'}
    >
      {picture ? <AvatarPicture avatar={avatar} size={size} /> : initialsOf(name)}
    </span>
  );
};

interface AvatarPickerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
  name?: string;
  current: ChosenAvatar | undefined;
  /**
   * The first-login welcome: a greeting instead of the settings title, and
   * "Skip for now" instead of "Use default". Closing it in any way counts as
   * the skip, so it is never shown twice.
   */
  welcome?: boolean;
  onSkip?: () => void;
}

/** Choose a photo, an illustration, an initials colour or an emoji. */
export const AvatarPicker: React.FC<AvatarPickerProps> = ({
  open,
  onOpenChange,
  userId,
  name,
  current,
  welcome = false,
  onSkip,
}) => {
  const { t } = useTranslation();
  const update = useUpdateAvatar();
  const [tab, setTab] = useState<Tab>(current?.kind ?? 'illustrated');
  const [draft, setDraft] = useState<ChosenAvatar | null>(current ?? null);
  const [style, setStyle] = useState(
    current?.kind === 'illustrated' && current.style ? current.style : AVATAR_STYLES[0].id
  );
  const [seeds, setSeeds] = useState(freshSeeds);
  const [photo, setPhoto] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const fileInput = useRef<HTMLInputElement>(null);
  const defaultColor = avatarColorIndex(userId);
  const colorOf = (a: ChosenAvatar | null): number =>
    a && (a.kind === 'initials' || a.kind === 'emoji') && a.color !== undefined
      ? a.color
      : defaultColor;

  const selectedStyle = useMemo(() => AVATAR_STYLES.find((s) => s.id === style), [style]);

  const save = (next: ChosenAvatar | null): void => {
    update.mutate(next, {
      onSuccess: () => {
        toast.success(t('settings.avatar.saved'));
        onOpenChange(false);
      },
      onError: () => toast.error(t('settings.avatar.saveFailed')),
    });
  };

  const pickFile = (file: File | undefined): void => {
    if (!file) return;
    if (!PHOTO_TYPES.includes(file.type)) {
      toast.error(t('settings.avatar.wrongType'));
      return;
    }
    if (file.size > MAX_PHOTO_BYTES) {
      toast.error(t('settings.avatar.tooLarge'));
      return;
    }
    setPhoto(file);
  };

  const uploadCropped = async (blob: Blob): Promise<void> => {
    setUploading(true);
    try {
      const asset = await uploadAsset(new File([blob], 'avatar.webp', { type: blob.type }), {
        visibility: 'public',
      });
      const next: ChosenAvatar = { kind: 'photo', url: asset.url };
      setDraft(next);
      setPhoto(null);
      save(next);
    } catch {
      toast.error(t('settings.avatar.uploadFailed'));
    } finally {
      setUploading(false);
    }
  };

  const swatches = (selected: number, onPick: (index: number) => void): React.ReactNode => (
    <div
      className="flex flex-wrap gap-2"
      role="radiogroup"
      aria-label={t('settings.avatar.colour')}
    >
      {AVATAR_PALETTE.map((hex, index) => (
        <button
          key={hex}
          type="button"
          role="radio"
          aria-checked={selected === index}
          aria-label={t('settings.avatar.colourN', { n: index + 1 })}
          onClick={() => onPick(index)}
          className="h-8 w-8 rounded-full"
          style={{
            background: hex,
            boxShadow:
              selected === index
                ? '0 0 0 2px var(--background, #fff), 0 0 0 4px currentColor'
                : undefined,
          }}
        />
      ))}
    </div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[640px]">
        <DialogHeader>
          <DialogTitle>
            {welcome ? t('settings.avatar.welcomeTitle') : t('settings.avatar.dialogTitle')}
          </DialogTitle>
          <DialogDescription>
            {welcome
              ? t('settings.avatar.welcomeDescription')
              : t('settings.avatar.dialogDescription')}
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-4">
          <AvatarPreview avatar={draft} userId={userId} name={name} size={72} />
          <div
            role="tablist"
            aria-label={t('settings.avatar.dialogTitle')}
            className="flex flex-wrap gap-1"
          >
            {TABS.map((id) => (
              <button
                key={id}
                type="button"
                role="tab"
                aria-selected={tab === id}
                onClick={() => setTab(id)}
                className={[
                  'min-h-9 rounded-md px-3 text-sm font-medium',
                  tab === id ? 'bg-accent/15 text-accent' : 'opacity-70 hover:opacity-100',
                ].join(' ')}
              >
                {t(`settings.avatar.tabs.${id}`)}
              </button>
            ))}
          </div>
        </div>

        <div role="tabpanel" className="min-h-[300px] pt-2">
          {tab === 'photo' &&
            (photo ? (
              <PhotoCropper
                file={photo}
                busy={uploading || update.isPending}
                onCropped={(blob) => void uploadCropped(blob)}
                onCancel={() => {
                  setPhoto(null);
                  fileInput.current?.click();
                }}
              />
            ) : (
              <button
                type="button"
                onClick={() => fileInput.current?.click()}
                onDragOver={(e) => e.preventDefault()}
                onDrop={(e) => {
                  e.preventDefault();
                  pickFile(e.dataTransfer.files[0]);
                }}
                className="flex h-[260px] w-full flex-col items-center justify-center gap-2 rounded-lg border border-dashed text-sm"
              >
                <Upload size={20} aria-hidden="true" />
                <span className="font-medium">{t('settings.avatar.uploadPhoto')}</span>
                <span className="opacity-70">{t('settings.avatar.photoHint')}</span>
              </button>
            ))}
          <input
            ref={fileInput}
            type="file"
            accept={PHOTO_TYPES.join(',')}
            className="hidden"
            onChange={(e) => {
              pickFile(e.target.files?.[0]);
              e.target.value = '';
            }}
          />

          {tab === 'illustrated' && (
            <div className="space-y-3">
              {AVATAR_STYLE_GROUPS.map((group) => (
                <div key={group} className="flex flex-wrap items-center gap-1.5">
                  <span className="w-24 text-xs font-medium opacity-70">
                    {t(`settings.avatar.groups.${group}`)}
                  </span>
                  {AVATAR_STYLES.filter((s) => s.group === group).map((s) => (
                    <button
                      key={s.id}
                      type="button"
                      aria-pressed={style === s.id}
                      onClick={() => setStyle(s.id)}
                      className={[
                        'min-h-8 rounded-full border px-2.5 text-xs',
                        style === s.id ? 'border-accent bg-accent/15 text-accent' : '',
                      ].join(' ')}
                    >
                      {s.label}
                      {s.isNew && (
                        <span className="ms-1 opacity-70">· {t('settings.avatar.new')}</span>
                      )}
                    </button>
                  ))}
                </div>
              ))}
              <div
                className="grid grid-cols-6 gap-2"
                role="listbox"
                aria-label={selectedStyle?.label}
              >
                {(style === MOODS_STYLE_ID ? MOODS.map((m) => m.id) : seeds).map((seed) => {
                  const option: ChosenAvatar = { kind: 'illustrated', style, seed };
                  const selected =
                    draft?.kind === 'illustrated' && draft.style === style && draft.seed === seed;
                  return (
                    <button
                      key={seed}
                      type="button"
                      role="option"
                      aria-selected={selected}
                      aria-label={
                        style === MOODS_STYLE_ID
                          ? (MOODS.find((m) => m.id === seed)?.label ?? seed)
                          : t('settings.avatar.optionN', { style: selectedStyle?.label })
                      }
                      onClick={() => setDraft(option)}
                      className="aspect-square overflow-hidden rounded-full"
                      style={{ boxShadow: selected ? '0 0 0 3px currentColor' : undefined }}
                    >
                      <AvatarPicture avatar={option} size={64} />
                    </button>
                  );
                })}
              </div>
              <div className="flex items-center justify-between gap-2">
                <button
                  type="button"
                  hidden={style === MOODS_STYLE_ID}
                  onClick={() => setSeeds(freshSeeds())}
                  className="inline-flex min-h-8 items-center gap-1.5 rounded-md border px-2.5 text-sm"
                >
                  <Shuffle size={14} aria-hidden="true" />
                  {t('settings.avatar.shuffle')}
                </button>
                {selectedStyle?.credit && (
                  <p className="text-xs opacity-70" data-avatar-credit>
                    <a
                      href={selectedStyle.credit.workUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="underline"
                    >
                      {selectedStyle.credit.work}
                    </a>{' '}
                    {t('settings.avatar.by', { author: selectedStyle.credit.author })} ·{' '}
                    <a
                      href={selectedStyle.credit.licenseUrl}
                      target="_blank"
                      rel="noreferrer"
                      className="underline"
                    >
                      {selectedStyle.credit.license}
                    </a>{' '}
                    · {t('settings.avatar.remix')}
                  </p>
                )}
              </div>
            </div>
          )}

          {tab === 'initials' && (
            <div className="space-y-3">
              <p className="text-sm opacity-70">{t('settings.avatar.initialsHint')}</p>
              {swatches(draft?.kind === 'initials' ? colorOf(draft) : -1, (index) =>
                setDraft({ kind: 'initials', color: index })
              )}
            </div>
          )}

          {tab === 'emoji' && (
            <div className="space-y-3">
              <div
                className="grid grid-cols-12 gap-1"
                role="listbox"
                aria-label={t('settings.avatar.tabs.emoji')}
              >
                {AVATAR_EMOJI.map((emoji) => {
                  const selected = draft?.kind === 'emoji' && draft.emoji === emoji;
                  return (
                    <button
                      key={emoji}
                      type="button"
                      role="option"
                      aria-selected={selected}
                      onClick={() => setDraft({ kind: 'emoji', emoji, color: colorOf(draft) })}
                      className={[
                        'h-9 rounded-md text-xl',
                        selected ? 'bg-accent/15' : 'hover:bg-accent/10',
                      ].join(' ')}
                    >
                      {emoji}
                    </button>
                  );
                })}
              </div>
              {swatches(draft?.kind === 'emoji' ? colorOf(draft) : -1, (index) =>
                setDraft({
                  kind: 'emoji',
                  emoji: draft?.kind === 'emoji' && draft.emoji ? draft.emoji : AVATAR_EMOJI[0],
                  color: index,
                })
              )}
            </div>
          )}
        </div>

        <div className="flex items-center justify-between gap-2 border-t pt-3">
          {welcome ? (
            <button
              type="button"
              onClick={() => {
                onSkip?.();
                onOpenChange(false);
              }}
              className="min-h-9 rounded-md px-3 text-sm opacity-80 hover:opacity-100"
            >
              {t('settings.avatar.skip')}
            </button>
          ) : (
            <button
              type="button"
              onClick={() => save(null)}
              disabled={update.isPending || !current}
              className="min-h-9 rounded-md px-3 text-sm opacity-80 hover:opacity-100 disabled:opacity-40"
            >
              {t('settings.avatar.useDefault')}
            </button>
          )}
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => onOpenChange(false)}
              className="min-h-9 rounded-md border px-3 text-sm"
            >
              {t('settings.avatar.cancel')}
            </button>
            <button
              type="button"
              onClick={() => save(draft)}
              disabled={update.isPending || !draft || tab === 'photo'}
              className="min-h-9 rounded-md bg-primary px-3 text-sm font-medium text-white disabled:opacity-50"
            >
              {update.isPending ? t('settings.avatar.saving') : t('settings.avatar.save')}
            </button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
};

export default AvatarPicker;
