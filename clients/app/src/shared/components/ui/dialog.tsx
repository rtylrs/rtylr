import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogTrigger,
  DialogClose,
  DialogContent as SharedDialogContent,
  DialogTitle,
  DialogDescription as SharedDialogDescription,
} from '@voxire/ui';
import { cn } from '../../lib/utils';

const DialogContent = React.forwardRef<
  React.ElementRef<typeof SharedDialogContent>,
  React.ComponentPropsWithoutRef<typeof SharedDialogContent>
>(({ className, ...props }, ref) => {
  const { t } = useTranslation();
  return (
    <SharedDialogContent
      ref={ref}
      closeLabel={t('common.closeDialog')}
      className={cn('portal-form-dialog', className)}
      {...props}
    />
  );
});
DialogContent.displayName = 'DialogContent';

const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn('flex flex-col gap-2 p-6 pb-4 pe-14', className)} {...props} />
);

const DialogDescription = React.forwardRef<
  React.ElementRef<typeof SharedDialogDescription>,
  React.ComponentPropsWithoutRef<typeof SharedDialogDescription>
>(({ className, ...props }, ref) => (
  <SharedDialogDescription
    ref={ref}
    className={cn('text-[var(--dash-text-dim)] leading-relaxed', className)}
    {...props}
  />
));
DialogDescription.displayName = 'DialogDescription';

const DialogFooter = ({ className, style, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={cn('flex flex-wrap items-center justify-end gap-2 p-6 pt-4', className)}
    style={{ borderTop: '1px solid var(--dash-border)', ...style }}
    {...props}
  />
);

export {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
};
