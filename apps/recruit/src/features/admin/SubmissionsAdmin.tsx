import React, { useState } from 'react';
import { toast } from 'sonner';
import { Mail, Phone, FileText } from 'lucide-react';

import {
  useAdminSubmissions,
  useUpdateSubmissionStatus,
  fetchResumeURL,
} from '../../shared/hooks/queries/useAdmin';

/**
 * SubmissionsAdmin — HR-facing inbox for all candidate applications.
 * Kept deliberately simple (no kanban, no stages swimlane) for MVP;
 * the status dropdown on each row is enough to move a candidate
 * through the funnel.
 */
const STATUSES = ['new', 'reviewed', 'interviewing', 'offered', 'hired', 'rejected'];

const SubmissionsAdmin: React.FC = (): React.JSX.Element => {
  const [filter, setFilter] = useState<string>('');
  const { data, isLoading } = useAdminSubmissions(filter);
  const update = useUpdateSubmissionStatus();
  const submissions = data?.submissions ?? [];

  return (
    <div className="px-10 py-8">
      <h1 className="text-2xl font-semibold tracking-tight">Submissions</h1>
      <p className="mt-1 text-sm text-white/60">
        {submissions.length} {submissions.length === 1 ? 'application' : 'applications'}
        {filter && ` in "${filter}"`}
      </p>

      <div className="mt-6 flex items-center gap-2 text-sm">
        <FilterChip active={filter === ''} onClick={() => setFilter('')}>
          All
        </FilterChip>
        {STATUSES.map((s) => (
          <FilterChip key={s} active={filter === s} onClick={() => setFilter(s)}>
            {s}
          </FilterChip>
        ))}
      </div>

      <div
        className="mt-6 overflow-hidden rounded-lg border"
        style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
      >
        {isLoading ? (
          <div className="p-10 text-center text-sm text-white/60">Loading…</div>
        ) : submissions.length === 0 ? (
          <div className="p-10 text-center text-sm text-white/60">
            No applications yet{filter && ` at this stage`}.
          </div>
        ) : (
          <ul className="divide-y" style={{ borderColor: '#1f1f1f' }}>
            {submissions.map((s) => (
              <li
                key={s.id}
                className="flex items-start gap-6 px-6 py-5 transition hover:bg-white/5"
              >
                <div className="min-w-0 flex-1">
                  <div className="flex items-baseline gap-3">
                    <div className="font-semibold">{s.candidateName}</div>
                    <div className="text-xs text-white/50">
                      {new Date(s.createdAt).toLocaleDateString()}
                    </div>
                  </div>
                  <div className="mt-0.5 text-sm text-white/70">
                    Applied for <span className="text-white">{s.careerTitle || '—'}</span>
                  </div>
                  <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-white/60">
                    <a
                      href={`mailto:${s.email}`}
                      className="inline-flex items-center gap-1 hover:text-white"
                    >
                      <Mail className="h-3 w-3" />
                      {s.email}
                    </a>
                    {s.phone && (
                      <span className="inline-flex items-center gap-1">
                        <Phone className="h-3 w-3" />
                        {s.phone}
                      </span>
                    )}
                    {s.resumeUrl && (
                      <button
                        onClick={async () => {
                          try {
                            const url = await fetchResumeURL(s.id);
                            window.open(url, '_blank', 'noopener,noreferrer');
                          } catch {
                            toast.error("Couldn't open resume");
                          }
                        }}
                        className="inline-flex items-center gap-1 hover:text-white"
                      >
                        <FileText className="h-3 w-3" />
                        Resume
                      </button>
                    )}
                  </div>
                  {s.coverLetter && (
                    <p className="mt-3 line-clamp-3 text-sm text-white/70">{s.coverLetter}</p>
                  )}
                </div>

                <select
                  value={s.status}
                  onChange={(e) =>
                    update.mutate(
                      { id: s.id, status: e.target.value },
                      {
                        onSuccess: () => toast.success('Status updated'),
                        onError: () => toast.error("Couldn't update"),
                      }
                    )
                  }
                  className="rounded-md border bg-black/40 px-3 py-1.5 text-sm outline-none focus:border-white/40"
                  style={{ borderColor: '#2a2a2a' }}
                >
                  {STATUSES.map((opt) => (
                    <option key={opt} value={opt}>
                      {opt}
                    </option>
                  ))}
                </select>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

const FilterChip: React.FC<{
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}> = ({ active, onClick, children }) => (
  <button
    onClick={onClick}
    className={`rounded-full px-3 py-1 text-xs font-medium transition ${
      active ? 'bg-white text-black' : 'border border-white/10 text-white/70 hover:bg-white/5'
    }`}
  >
    {children}
  </button>
);

export default SubmissionsAdmin;
