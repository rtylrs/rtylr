import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from '@tanstack/react-router';
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../../shared/components/ui';
import { useCreateGroup } from '../../shared/hooks/queries/useGroups';
import { buildPermissionProfileAssignments } from './authorizationDraft';
import { DEFAULT_GROUP_COLOR, GroupAppAccess, GroupGeneralFields } from './GroupForm';

interface NewGroupDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

/** Creates a group, then opens it so members can be added. */
export function NewGroupDialog({ open, onOpenChange }: NewGroupDialogProps): React.JSX.Element {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const createGroup = useCreateGroup();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [color, setColor] = useState(DEFAULT_GROUP_COLOR);
  const [licensedAppIds, setLicensedAppIds] = useState<string[]>([]);
  const [profileIdsByApp, setProfileIdsByApp] = useState<Record<string, string[]>>({});

  const reset = (): void => {
    setName('');
    setDescription('');
    setColor(DEFAULT_GROUP_COLOR);
    setLicensedAppIds([]);
    setProfileIdsByApp({});
  };

  const handleOpenChange = (next: boolean): void => {
    if (!next) reset();
    onOpenChange(next);
  };

  const handleSubmit = async (event: React.FormEvent): Promise<void> => {
    event.preventDefault();
    if (!name.trim()) return;
    try {
      const created = await createGroup.mutateAsync({
        name: name.trim(),
        description,
        color,
        licensedAppIds,
        permissionProfileAssignments: buildPermissionProfileAssignments(
          licensedAppIds,
          profileIdsByApp,
          'group'
        ),
      });
      handleOpenChange(false);
      void navigate({ to: '/team/groups/$id', params: { id: String(created.id) } });
    } catch {
      // onError in the hook shows the toast; keep the dialog open to retry.
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-3xl" aria-describedby={undefined}>
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{t('team.groupDetail.newGroupTitle')}</DialogTitle>
          </DialogHeader>
          <div className="grid gap-5 px-6 pb-6">
            <GroupGeneralFields
              name={name}
              description={description}
              color={color}
              onNameChange={setName}
              onDescriptionChange={setDescription}
              onColorChange={setColor}
            />
            <div
              className="overflow-x-auto rounded-xl"
              style={{ border: '1px solid var(--dash-border)' }}
            >
              <GroupAppAccess
                accentColor={color}
                licensedAppIds={licensedAppIds}
                profileIdsByApp={profileIdsByApp}
                onToggleApp={(appId) =>
                  setLicensedAppIds((current) =>
                    current.includes(appId)
                      ? current.filter((id) => id !== appId)
                      : [...current, appId]
                  )
                }
                onProfilesChange={(appId, ids) =>
                  setProfileIdsByApp((current) => ({ ...current, [appId]: ids }))
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={createGroup.isPending || !name.trim()}>
              {createGroup.isPending
                ? t('team.groupDetail.savingButton')
                : t('team.groups.newGroup')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
