import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { createLocationSchema } from '../../shared/schemas/locations';
import { useCreateLocation } from '../../shared/hooks/queries/useLocations';
import { Modal, Button, Input, Label } from '../../shared/components/ui';

interface CreateLocationFormProps {
  open: boolean;
  onClose: () => void;
  onCreated: (location: { id: number; name: string }) => void;
}

const CreateLocationForm = ({
  open,
  onClose,
  onCreated,
}: CreateLocationFormProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const createLocation = useCreateLocation();

  const [name, setName] = useState('');
  const [address, setAddress] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();

    const result = createLocationSchema.safeParse({ name, address });
    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = String(issue.path[0]);
        fieldErrors[key] = issue.message;
      }
      setErrors(fieldErrors);
      return;
    }

    setErrors({});
    createLocation.mutate(
      {
        name: result.data.name,
        address: result.data.address || undefined,
      },
      {
        onSuccess: (res) => {
          if (res.location?.id) {
            onCreated({ id: Number(res.location.id), name: result.data.name });
          }
          setName('');
          setAddress('');
          onClose();
        },
        onError: () => toast.error(t('errors.generic')),
      }
    );
  };

  return (
    <Modal open={open} onClose={onClose} title={t('employees.form.createLocation.title')}>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div>
          <Label htmlFor="location-name">{t('employees.form.createLocation.name')}</Label>
          <Input
            id="location-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('employees.form.createLocation.namePlaceholder')}
          />
          {errors.name && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
              {t('common.required')}
            </p>
          )}
        </div>

        <div>
          <Label htmlFor="location-address">{t('employees.form.createLocation.address')}</Label>
          <Input
            id="location-address"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            placeholder={t('employees.form.createLocation.addressPlaceholder')}
          />
        </div>

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
          <Button type="button" variant="ghost" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" variant="primary" disabled={createLocation.isPending}>
            {createLocation.isPending
              ? t('common.loading')
              : t('employees.form.createLocation.submit')}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default CreateLocationForm;
