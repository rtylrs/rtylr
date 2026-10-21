import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { ArrowDown, ArrowUp, Plus, Trash2 } from 'lucide-react';
import { Breadcrumbs, Button, ConfirmModal } from '../../shared/components/ui';
import type { Pipeline } from './data/model';
import { useDeletePipeline, usePipelines, useSavePipeline } from './data/queries';
import { LoadFailed } from './components/common';

interface DraftStage {
  key: string;
  id?: string;
  name: string;
  probability: number;
  kind: 'open' | 'won' | 'lost';
}

let draftKey = 0;
const nextKey = () => `stage-${(draftKey += 1)}`;

function draftOf(pipeline?: Pipeline): { name: string; stages: DraftStage[] } {
  return {
    name: pipeline?.name ?? '',
    stages: (pipeline?.stages ?? []).map((stage) => ({
      key: stage.id ?? nextKey(),
      id: stage.id,
      name: stage.name ?? '',
      probability: stage.probability ?? 0,
      kind: (stage.kind as DraftStage['kind']) ?? 'open',
    })),
  };
}

/** Edits a pipeline; without one it creates a new pipeline from `seed`. */
const PipelineEditor: React.FC<{ pipeline?: Pipeline; seed?: Pipeline; onDone?: () => void }> = ({
  pipeline,
  seed,
  onDone,
}) => {
  const { t } = useTranslation();
  const save = useSavePipeline();
  const remove = useDeletePipeline();
  const [draft, setDraft] = useState(() => draftOf(pipeline ?? seed));
  const [deleting, setDeleting] = useState(false);
  useEffect(() => {
    if (pipeline) setDraft(draftOf(pipeline));
  }, [pipeline]);

  const openStages = draft.stages.filter((stage) => stage.kind === 'open');
  const closing = draft.stages.filter((stage) => stage.kind !== 'open');
  const update = (key: string, change: Partial<DraftStage>) =>
    setDraft((current) => ({
      ...current,
      stages: current.stages.map((stage) => (stage.key === key ? { ...stage, ...change } : stage)),
    }));
  const moveStage = (key: string, by: number) =>
    setDraft((current) => {
      const opens = current.stages.filter((stage) => stage.kind === 'open');
      const index = opens.findIndex((stage) => stage.key === key);
      const target = index + by;
      if (index < 0 || target < 0 || target >= opens.length) return current;
      const reordered = [...opens];
      [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
      return {
        ...current,
        stages: [...reordered, ...current.stages.filter((stage) => stage.kind !== 'open')],
      };
    });
  const valid =
    draft.name.trim() !== '' &&
    openStages.length > 0 &&
    draft.stages.every((stage) => stage.name.trim() !== '');

  const submit = (): void => {
    save.mutate(
      {
        id: pipeline?.id,
        pipeline: {
          name: draft.name.trim(),
          stages: [...openStages, ...closing].map((stage) => ({
            id: stage.id,
            name: stage.name.trim(),
            probability:
              stage.kind === 'open'
                ? Math.min(100, Math.max(0, Math.round(stage.probability)))
                : undefined,
            kind: stage.kind,
          })),
        },
      },
      { onSuccess: onDone }
    );
  };

  return (
    <section className="sales-panel" aria-label={draft.name || t('crmSales.pipelines.new')}>
      <div className="sales-panel-head">
        <input
          className="sales-input"
          style={{ fontWeight: 600, flex: 1, maxWidth: 360 }}
          aria-label={t('crmSales.pipelines.name')}
          placeholder={t('crmSales.pipelines.name')}
          value={draft.name}
          onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))}
        />
        {pipeline?.isDefault && (
          <span className="sales-badge">{t('crmSales.pipelines.default')}</span>
        )}
      </div>
      <table className="sales-table" style={{ cursor: 'default' }}>
        <thead>
          <tr>
            <th>{t('crmSales.pipelines.stage')}</th>
            <th>{t('crmSales.pipelines.probability')}</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {[...openStages, ...closing].map((stage) => {
            const index = openStages.findIndex((item) => item.key === stage.key);
            return (
              <tr key={stage.key} style={{ cursor: 'default' }}>
                <td>
                  <input
                    className="sales-input"
                    style={{ width: '100%' }}
                    aria-label={t('crmSales.pipelines.stageName')}
                    value={stage.name}
                    onChange={(event) => update(stage.key, { name: event.target.value })}
                  />
                </td>
                <td style={{ width: 140 }}>
                  {stage.kind === 'open' ? (
                    <label style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}>
                      <input
                        className="sales-input"
                        style={{ width: 72 }}
                        type="number"
                        min={0}
                        max={100}
                        aria-label={t('crmSales.pipelines.probability')}
                        value={stage.probability}
                        onChange={(event) =>
                          update(stage.key, { probability: Number(event.target.value) })
                        }
                      />
                      %
                    </label>
                  ) : (
                    <span
                      className="sales-badge"
                      data-tone={stage.kind === 'won' ? 'green' : 'red'}
                    >
                      {t(`crmSales.pipelines.kind.${stage.kind}`)}
                    </span>
                  )}
                </td>
                <td style={{ width: 120, textAlign: 'end' }}>
                  {stage.kind === 'open' && (
                    <span className="sales-chip-row" style={{ justifyContent: 'flex-end' }}>
                      <button
                        type="button"
                        className="sales-chip"
                        style={{ minHeight: 28, padding: 4 }}
                        aria-label={t('crmSales.pipelines.moveUp')}
                        disabled={index === 0}
                        onClick={() => moveStage(stage.key, -1)}
                      >
                        <ArrowUp size={12} />
                      </button>
                      <button
                        type="button"
                        className="sales-chip"
                        style={{ minHeight: 28, padding: 4 }}
                        aria-label={t('crmSales.pipelines.moveDown')}
                        disabled={index === openStages.length - 1}
                        onClick={() => moveStage(stage.key, 1)}
                      >
                        <ArrowDown size={12} />
                      </button>
                      <button
                        type="button"
                        className="sales-chip"
                        style={{ minHeight: 28, padding: 4 }}
                        aria-label={t('crmSales.pipelines.removeStage')}
                        disabled={openStages.length === 1}
                        onClick={() =>
                          setDraft((current) => ({
                            ...current,
                            stages: current.stages.filter((item) => item.key !== stage.key),
                          }))
                        }
                      >
                        <Trash2 size={12} />
                      </button>
                    </span>
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      <div className="sales-filters">
        <Button
          variant="outline"
          onClick={() =>
            setDraft((current) => ({
              ...current,
              stages: [
                ...current.stages.filter((stage) => stage.kind === 'open'),
                { key: nextKey(), name: '', probability: 50, kind: 'open' },
                ...current.stages.filter((stage) => stage.kind !== 'open'),
              ],
            }))
          }
        >
          <Plus size={13} />
          {t('crmSales.pipelines.addStage')}
        </Button>
        <span className="sales-spacer" />
        {pipeline && !pipeline.isDefault && (
          <Button variant="ghost" onClick={() => setDeleting(true)}>
            <Trash2 size={13} />
            {t('crmSales.pipelines.delete')}
          </Button>
        )}
        {onDone && !pipeline && (
          <Button variant="ghost" onClick={onDone}>
            {t('crmSales.common.cancel')}
          </Button>
        )}
        <Button onClick={submit} disabled={!valid || save.isPending}>
          {save.isPending ? t('crmSales.common.saving') : t('crmSales.common.save')}
        </Button>
      </div>
      <ConfirmModal
        open={deleting}
        onClose={() => setDeleting(false)}
        onConfirm={() => pipeline?.id && remove.mutate(pipeline.id)}
        title={t('crmSales.pipelines.deleteTitle')}
        description={t('crmSales.pipelines.deleteHint')}
        confirmLabel={t('crmSales.common.delete')}
        variant="danger"
      />
    </section>
  );
};

const TEMPLATE: Pipeline = {
  stages: [
    { name: 'Qualified', probability: 10, kind: 'open' },
    { name: 'Proposal', probability: 50, kind: 'open' },
    { name: 'Won', probability: 100, kind: 'won' },
    { name: 'Lost', probability: 0, kind: 'lost' },
  ],
};

const Pipelines: React.FC = () => {
  const { t } = useTranslation();
  const pipelines = usePipelines();
  const [adding, setAdding] = useState(false);
  const template: Pipeline = {
    stages: TEMPLATE.stages?.map((stage) => ({
      ...stage,
      name: t(
        `crmSales.pipelines.template.${stage.kind === 'open' ? (stage.probability === 10 ? 'qualified' : 'proposal') : stage.kind}`
      ),
    })),
  };
  return (
    <div className="sales-page" style={{ maxWidth: 820 }}>
      <HelmetContainer title={`${t('crmSales.pipelines.title')} · rtylr`} />
      <Breadcrumbs
        items={[
          { label: t('crmSales.deals.title'), to: '/crm/sales' },
          { label: t('crmSales.pipelines.title') },
        ]}
      />
      <header className="sales-header">
        <div>
          <h1>{t('crmSales.pipelines.title')}</h1>
          <p>{t('crmSales.pipelines.subtitle')}</p>
        </div>
        {!adding && (
          <Button variant="outline" onClick={() => setAdding(true)}>
            <Plus size={13} />
            {t('crmSales.pipelines.add')}
          </Button>
        )}
      </header>
      {pipelines.isError ? (
        <LoadFailed
          title={t('crmSales.pipelines.loadFailed')}
          onRetry={() => void pipelines.refetch()}
        />
      ) : (
        <>
          {(pipelines.data ?? []).map((pipeline) => (
            <PipelineEditor key={pipeline.id} pipeline={pipeline} />
          ))}
          {adding && <PipelineEditor key="new" seed={template} onDone={() => setAdding(false)} />}
        </>
      )}
    </div>
  );
};

export default Pipelines;
