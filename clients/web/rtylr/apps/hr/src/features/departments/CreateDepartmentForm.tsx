import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { createDepartmentSchema } from '../../shared/schemas/departments';
import { useCreateDepartment } from './useDepartments';
import { Modal, Button, Input, Label, Textarea } from '../../shared/components/ui';

interface CreateDepartmentFormProps {
  open: boolean;
  onClose: () => void;
}

const CreateDepartmentForm = ({
  open,
  onClose,
}: CreateDepartmentFormProps): React.JSX.Element | null => {
  const { t } = useTranslation();
  const createDepartment = useCreateDepartment();

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();

    const result = createDepartmentSchema.safeParse({ name, description });
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
    createDepartment.mutate(
      {
        name: result.data.name,
        description: result.data.description || undefined,
      },
      {
        onSuccess: () => {
          setName('');
          setDescription('');
          toast.success(t('departments.created'));
          onClose();
        },
        onError: () => toast.error(t('errors.generic')),
      }
    );
  };

  return (
    <Modal open={open} onClose={onClose} title={t('departments.form.title')}>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div>
          <Label htmlFor="department-name">{t('departments.form.name')}</Label>
          <Input
            id="department-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder={t('departments.form.namePlaceholder')}
          />
          {errors.name && (
            <p style={{ color: 'var(--hr-error)', fontSize: 11, marginTop: 4 }}>
              {t('common.required')}
            </p>
          )}
        </div>

        <div>
          <Label htmlFor="department-description">{t('departments.form.description')}</Label>
          <Textarea
            id="department-description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('departments.form.descriptionPlaceholder')}
          />
        </div>

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
          <Button type="button" variant="ghost" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button type="submit" variant="primary" disabled={createDepartment.isPending}>
            {createDepartment.isPending ? t('common.loading') : t('departments.form.submit')}
          </Button>
        </div>
      </form>
    </Modal>
  );
};

export default CreateDepartmentForm;
