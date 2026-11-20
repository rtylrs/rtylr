import React, { useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Shuffle, Upload } from 'lucide-react';
import { uploadAsset } from '@voxire/api';
import { AVATAR_PALETTE, avatarColor, avatarColorAt, avatarColorIndex } from '@voxire/ui';
import { Button } from '../components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../components/ui/dialog';
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
  framed?: boolean;
}> = ({ avatar, userId, name, size, framed = false }) => {
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
      {picture ? <AvatarPicture avatar={avatar} size={size} framed={framed} /> : initialsOf(name)}
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
  const framedStyle = (id: string | undefined): boolean =>
    !AVATAR_STYLES.find((s) => s.id === id)?.fullBleed;
  const draftLabel = !draft
    ? t('settings.avatar.previewDefault')
    : draft.kind === 'illustrated'
      ? `${t('settings.avatar.tabs.illustrated')} · ${
          AVATAR_STYLES.find((s) => s.id === draft.style)?.label ?? ''
        }`
      : t(`settings.avatar.tabs.${draft.kind}`);

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
          className="h-9 w-9 rounded-full transition-shadow"
          style={{
            background: hex,
            boxShadow:
              selected === index
                ? '0 0 0 2px var(--dash-surface), 0 0 0 4px var(--dash-text)'
                : undefined,
          }}
        />
      ))}
    </div>
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent
        className="max-w-[680px] outline-none"
        // Start on the dialog rather than ringing the first tab; Tab moves into the tabs.
        onOpenAutoFocus={(event) => {
          event.preventDefault();
          (event.currentTarget as HTMLElement).focus();
        }}
      >
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

        <div className="space-y-4 px-6">
          <div className="flex items-center gap-4">
            <AvatarPreview
              avatar={draft}
              userId={userId}
              name={name}
              size={64}
              framed={draft?.kind === 'illustrated' && framedStyle(draft.style)}
            />
            <div className="min-w-0">
              {name && <p className="truncate text-sm font-semibold">{name}</p>}
              <p className="truncate text-[13px] text-[var(--dash-text-dim)]">{draftLabel}</p>
            </div>
          </div>

          <div
            role="tablist"
            aria-label={t('settings.avatar.dialogTitle')}
            className="flex gap-1 rounded-xl bg-[var(--dash-hover)] p-1"
          >
            {TABS.map((id) => (
              <button
                key={id}
                type="button"
                role="tab"
                aria-selected={tab === id}
                onClick={() => setTab(id)}
                className={[
                  'min-h-9 flex-auto rounded-lg px-2.5 text-[13px] font-medium whitespace-nowrap transition-colors sm:flex-1',
                  tab === id
                    ? 'bg-[var(--dash-surface)] text-[var(--dash-text)] shadow-[0_1px_2px_rgb(0_0_0/0.08),0_0_0_1px_var(--dash-border)]'
                    : 'text-[var(--dash-text-dim)] hover:text-[var(--dash-text)]',
                ].join(' ')}
              >
                {t(`settings.avatar.tabs.${id}`)}
              </button>
            ))}
          </div>
        </div>

        <div role="tabpanel" className="px-6 pb-6 pt-4">
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
                className="flex h-[240px] w-full flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-[color-mix(in_srgb,var(--dash-text)_20%,transparent)] px-6 text-center text-sm transition-colors hover:bg-[var(--dash-hover)]"
              >
                <span className="mb-1 grid h-10 w-10 place-items-center rounded-full bg-[var(--dash-hover)]">
                  <Upload size={18} aria-hidden="true" />
                </span>
                <span className="font-medium">{t('settings.avatar.uploadPhoto')}</span>
                <span className="text-[13px] text-[var(--dash-text-dim)]">
                  {t('settings.avatar.photoHint')}
                </span>
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
            <div className="space-y-5">
              <div className="grid gap-x-3 gap-y-3 sm:grid-cols-[84px_1fr]">
                {AVATAR_STYLE_GROUPS.map((group) => (
                  <React.Fragment key={group}>
                    <span className="text-xs font-medium text-[var(--dash-text-dim)] sm:pt-1.5">
                      {t(`settings.avatar.groups.${group}`)}
                    </span>
                    <div className="-mt-1.5 flex flex-wrap gap-1.5 sm:mt-0">
                      {AVATAR_STYLES.filter((s) => s.group === group).map((s) => (
                        <button
                          key={s.id}
                          type="button"
                          aria-pressed={style === s.id}
                          onClick={() => setStyle(s.id)}
                          className={[
                            'inline-flex min-h-7 items-center gap-1.5 rounded-full border px-2.5 text-xs font-medium transition-colors',
                            style === s.id
                              ? 'border-transparent bg-[var(--dash-text)] text-[var(--dash-surface)]'
                              : 'border-[var(--dash-border)] text-[var(--dash-text-dim)] hover:bg-[var(--dash-hover)] hover:text-[var(--dash-text)]',
                          ].join(' ')}
                        >
                          {s.label}
                          {s.isNew && (
                            <span
                              className="h-1.5 w-1.5 rounded-full bg-accent"
                              title={t('settings.avatar.new')}
                            >
                              <span className="sr-only">{t('settings.avatar.new')}</span>
                            </span>
                          )}
                        </button>
                      ))}
                    </div>
                  </React.Fragment>
                ))}
              </div>

              <div
                className="grid grid-cols-4 gap-3 sm:grid-cols-6"
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
                      className={[
                        'aspect-square overflow-hidden rounded-full transition-shadow',
                        selected
                          ? 'shadow-[0_0_0_2px_var(--dash-surface),0_0_0_4px_var(--dash-text)]'
                          : 'hover:shadow-[0_0_0_2px_var(--dash-surface),0_0_0_4px_var(--dash-border)]',
                      ].join(' ')}
                    >
                      <AvatarPicture avatar={option} size={72} framed={framedStyle(style)} />
                    </button>
                  );
                })}
              </div>

              <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
                {style !== MOODS_STYLE_ID && (
                  <Button variant="outline" size="sm" onClick={() => setSeeds(freshSeeds())}>
                    <Shuffle aria-hidden="true" />
                    {t('settings.avatar.shuffle')}
                  </Button>
                )}
                {selectedStyle?.credit && (
                  <p
                    className="ms-auto text-xs text-[var(--dash-text-muted)] [&_a]:underline [&_a]:underline-offset-2 [&_a:hover]:text-[var(--dash-text)]"
                    data-avatar-credit
                  >
                    <a href={selectedStyle.credit.workUrl} target="_blank" rel="noreferrer">
                      {selectedStyle.credit.work}
                    </a>{' '}
                    {t('settings.avatar.by', { author: selectedStyle.credit.author })} ·{' '}
                    <a href={selectedStyle.credit.licenseUrl} target="_blank" rel="noreferrer">
                      {selectedStyle.credit.license}
                    </a>{' '}
                    · {t('settings.avatar.remix')}
                  </p>
                )}
              </div>
            </div>
          )}

          {tab === 'initials' && (
            <div className="space-y-4">
              <p className="text-sm text-[var(--dash-text-dim)]">
                {t('settings.avatar.initialsHint')}
              </p>
              {swatches(draft?.kind === 'initials' ? colorOf(draft) : -1, (index) =>
                setDraft({ kind: 'initials', color: index })
              )}
            </div>
          )}

          {tab === 'emoji' && (
            <div className="space-y-5">
              <div
                className="grid grid-cols-8 gap-1 sm:grid-cols-12"
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
                        'aspect-square rounded-lg text-xl transition-colors',
                        selected
                          ? 'bg-[var(--dash-active)] shadow-[inset_0_0_0_2px_var(--dash-text)]'
                          : 'hover:bg-[var(--dash-hover)]',
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

        <DialogFooter className="sticky bottom-0 justify-between bg-[var(--dash-surface)]">
          {welcome ? (
            <Button
              variant="ghost"
              onClick={() => {
                onSkip?.();
                onOpenChange(false);
              }}
            >
              {t('settings.avatar.skip')}
            </Button>
          ) : (
            <Button
              variant="ghost"
              onClick={() => save(null)}
              disabled={update.isPending || !current}
            >
              {t('settings.avatar.useDefault')}
            </Button>
          )}
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t('settings.avatar.cancel')}
            </Button>
            <Button
              onClick={() => save(draft)}
              disabled={update.isPending || !draft || tab === 'photo'}
            >
              {update.isPending ? t('settings.avatar.saving') : t('settings.avatar.save')}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export default AvatarPicker;
