import React, { useState } from 'react';
import { Trash2, UsersRound } from 'lucide-react';
import { HelmetContainer } from '@voxire/components';
import {
  useCreateTalentPool,
  useDeleteTalentPool,
  useRemoveTalentPoolMember,
  useTalentPoolMembers,
  useTalentPools,
} from '../../shared/hooks/queries/useAdmin';

const TalentPoolsAdmin: React.FC = (): React.JSX.Element => {
  const { data, isLoading } = useTalentPools();
  const create = useCreateTalentPool();
  const remove = useDeleteTalentPool();
  const [selectedPoolId, setSelectedPoolId] = useState<number | undefined>();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const members = useTalentPoolMembers(selectedPoolId);
  const removeMember = useRemoveTalentPoolMember();

  return (
    <>
      <HelmetContainer title="Talent pools · rtylr" />
      <div className="px-10 py-8">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Talent pools</h1>
          <p className="mt-1 text-sm text-white/60">
            Keep strong candidates organized for future openings and recurring hiring needs.
          </p>
        </div>

        <form
          className="mt-6 grid gap-2 rounded-lg border p-4 md:grid-cols-[minmax(0,220px)_minmax(0,1fr)_auto]"
          style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
          onSubmit={(event) => {
            event.preventDefault();
            if (!name.trim()) return;
            create.mutate(
              { name: name.trim(), description: description.trim() },
              { onSuccess: () => { setName(''); setDescription(''); } }
            );
          }}
        >
          <input value={name} onChange={(event) => setName(event.target.value)} maxLength={120} placeholder="Pool name" className="rounded border bg-black/30 px-3 py-2 text-sm text-white" />
          <input value={description} onChange={(event) => setDescription(event.target.value)} maxLength={500} placeholder="Description (optional)" className="rounded border bg-black/30 px-3 py-2 text-sm text-white" />
          <button type="submit" disabled={create.isPending || !name.trim()} className="rounded bg-white px-4 py-2 text-sm font-semibold text-black disabled:opacity-40">{create.isPending ? 'Creating…' : 'Create pool'}</button>
        </form>

        <div className="mt-6 grid gap-6 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)]">
          <section className="space-y-2" aria-label="Talent pools">
            {isLoading ? <p className="text-sm text-white/50">Loading pools…</p> : null}
            {(data?.pools ?? []).map((pool) => (
              <div
                key={pool.id}
                className="flex w-full items-center gap-3 rounded-lg border p-4 text-left"
                style={{ borderColor: selectedPoolId === pool.id ? '#fff' : '#1f1f1f', background: '#0f0f0f' }}
              >
                <button type="button" onClick={() => setSelectedPoolId(pool.id)} className="min-w-0 flex-1 text-left">
                  <span className="block truncate text-sm font-semibold">{pool.name}</span>
                  <span className="mt-1 block truncate text-xs text-white/50">{pool.description || 'No description'}</span>
                </button>
                <span className="shrink-0 text-xs text-white/45">{pool.memberCount} members</span>
                <button
                  type="button"
                  aria-label={`Delete ${pool.name}`}
                  onClick={() => remove.mutate(pool.id)}
                  className="shrink-0 rounded p-1 text-white/35 hover:text-red-300"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>
            ))}
            {!isLoading && (data?.pools ?? []).length === 0 && <p className="rounded-lg border p-6 text-sm text-white/45" style={{ borderColor: '#1f1f1f' }}>No talent pools yet.</p>}
          </section>

          <section className="rounded-lg border p-5" style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }} aria-label="Talent pool members">
            {!selectedPoolId ? (
              <div className="flex min-h-48 flex-col items-center justify-center text-center text-white/45">
                <UsersRound className="mb-3 h-6 w-6" />
                <p className="text-sm">Select a pool to view its candidates.</p>
              </div>
            ) : (
              <>
                <h2 className="text-lg font-semibold">Pool members</h2>
                <div className="mt-4 space-y-2">
                  {(members.data?.members ?? []).map((member) => (
                    <div key={member.id} className="flex items-center justify-between gap-3 rounded border p-3" style={{ borderColor: '#242424' }}>
                      <div className="min-w-0">
                        <div className="truncate text-sm font-medium">{member.candidateName}</div>
                        <div className="truncate text-xs text-white/50">{member.email} · {member.careerTitle || 'Open role'}</div>
                      </div>
                      <button type="button" className="text-xs text-white/45 hover:text-red-300" onClick={() => removeMember.mutate({ poolId: selectedPoolId, submissionId: member.submissionId })}>Remove</button>
                    </div>
                  ))}
                  {!members.isLoading && (members.data?.members ?? []).length === 0 && <p className="text-sm text-white/45">No candidates in this pool yet.</p>}
                </div>
              </>
            )}
          </section>
        </div>
      </div>
    </>
  );
};

export default TalentPoolsAdmin;
