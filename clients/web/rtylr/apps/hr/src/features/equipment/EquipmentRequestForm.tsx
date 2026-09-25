import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { z } from 'zod';
import { useCreateEquipmentRequest } from '../../shared/hooks/queries/useEquipment';
import { useMyEmployeeId } from '../../shared/hooks/useHrExtras';
import { Modal, Input, Textarea, Label, Button } from '../../shared/components/ui';

const equipmentRequestSchema = z.object({
  equipmentType: z.string().min(1),
  description: z.string().min(1),
  neededByDate: z.string().optional(),
});

interface Props {
  onClose: () => void;
}

const EquipmentRequestForm = ({ onClose }: Props): React.JSX.Element => {
  const { t } = useTranslation();
  const { data: myEmployeeId } = useMyEmployeeId();
  const employeeId = myEmployeeId !== undefined ? String(myEmployeeId) : undefined;
  const createRequest = useCreateEquipmentRequest();

  const [equipmentType, setEquipmentType] = useState('');
  const [description, setDescription] = useState('');
  const [neededByDate, setNeededByDate] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const handleSubmit = (): void => {
    const result = equipmentRequestSchema.safeParse({
      equipmentType,
      description,
      neededByDate: neededByDate || undefined,
    });

    if (!result.success) {
      const fieldErrors: Record<string, string> = {};
      for (const issue of result.error.issues) {
        const key = issue.path[0] as string;
        fieldErrors[key] = t('common.required');
      }
      setErrors(fieldErrors);
      return;
    }

    createRequest.mutate(
      {
        employeeId: employeeId ?? '',
        ...result.data,
        neededByDate: result.data.neededByDate ?? undefined,
      },
      {
        onSuccess: () => onClose(),
      }
    );
  };

  return (
    <Modal open onClose={onClose} title={t('equipment.form.title')} width={480}>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        <div>
          <Label>{t('equipment.form.type')}</Label>
          <Input
            value={equipmentType}
            onChange={(e) => setEquipmentType(e.target.value)}
            placeholder={t('equipment.form.typePlaceholder')}
            style={{ fontSize: 16 }}
          />
          {errors.equipmentType && (
            <span style={{ fontSize: 11, color: 'var(--hr-error)', marginTop: 2 }}>
              {errors.equipmentType}
            </span>
          )}
        </div>

        <div>
          <Label>{t('equipment.form.description')}</Label>
          <Textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder={t('equipment.form.descriptionPlaceholder')}
            style={{ fontSize: 16 }}
          />
          {errors.description && (
            <span style={{ fontSize: 11, color: 'var(--hr-error)', marginTop: 2 }}>
              {errors.description}
            </span>
          )}
        </div>

        <div>
          <Label>{t('equipment.form.neededBy')}</Label>
          <Input
            type="date"
            value={neededByDate}
            onChange={(e) => setNeededByDate(e.target.value)}
            style={{ fontSize: 16 }}
          />
        </div>

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end', marginTop: 4 }}>
          <Button variant="ghost" onClick={onClose}>
            {t('common.cancel')}
          </Button>
          <Button
            variant="primary"
            onClick={handleSubmit}
            disabled={createRequest.isPending || !employeeId}
          >
            {t('equipment.form.submit')}
          </Button>
        </div>
      </div>
    </Modal>
  );
};

export default EquipmentRequestForm;
