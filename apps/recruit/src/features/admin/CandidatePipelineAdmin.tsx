import React, { useMemo, useState } from 'react';
import { CalendarDays, FileText, Mail, Phone, Search, StickyNote, Tag, Users, X } from 'lucide-react';
import { toast } from 'sonner';
import { HelmetContainer } from '@voxire/components';
import {
  fetchResumeURL,
  useCreateSubmissionInterview,
  useCreateSubmissionNote,
  useCreateSubmissionOffer,
  useCreateSubmissionTag,
  useAdminSubmissions,
  useDeleteSubmissionTag,
  useSubmissionInterviews,
  useSubmissionEvents,
  useSubmissionNotes,
  useSubmissionOffers,
  useSubmissionTags,
  useAddTalentPoolMember,
  useTalentPools,
  useUpdateSubmissionOfferStatus,
  useUpdateSubmissionStatus,
} from '../../shared/hooks/queries/useAdmin';
import type { Submission, SubmissionEvent, SubmissionTag, TalentPool } from '../../shared/lib/types';

const PIPELINE = ['new', 'reviewed', 'interviewing', 'offered', 'hired', 'rejected'] as const;
type PipelineStatus = (typeof PIPELINE)[number];

const LABELS: Record<PipelineStatus, string> = {
  new: 'New',
  reviewed: 'Reviewed',
  interviewing: 'Interviewing',
  offered: 'Offer',
  hired: 'Hired',
  rejected: 'Rejected',
};

const COLORS: Record<PipelineStatus, string> = {
  new: '#60a5fa',
  reviewed: '#fbbf24',
  interviewing: '#c084fc',
  offered: '#fb923c',
  hired: '#34d399',
  rejected: '#f87171',
};

const CandidatePipelineAdmin: React.FC = (): React.JSX.Element => {
  const { data, isLoading } = useAdminSubmissions();
  const update = useUpdateSubmissionStatus();
  const [search, setSearch] = useState('');
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const submissions = data?.submissions ?? [];

  const filtered = useMemo(() => {
    const needle = search.trim().toLowerCase();
    if (!needle) return submissions;
    return submissions.filter((submission) =>
      [submission.candidateName, submission.email, submission.careerTitle, submission.status]
        .filter(Boolean)
        .some((value) => value.toLowerCase().includes(needle))
    );
  }, [search, submissions]);

  const selected = submissions.find((submission) => submission.id === selectedId) ?? null;

  const moveCandidate = (submission: Submission, status: string): void => {
    if (submission.status === status) return;
    update.mutate(
      { id: submission.id, status },
      {
        onSuccess: () => toast.success(`${submission.candidateName} moved to ${status}`),
        onError: () => toast.error("Couldn't update candidate stage"),
      }
    );
  };

  return (
    <>
      <HelmetContainer title="Candidate pipeline · rtylr" />
      <div className="px-10 py-8">
        <div className="flex flex-wrap items-end justify-between gap-4">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Candidate pipeline</h1>
            <p className="mt-1 text-sm text-white/60">
              Screen applicants, move them through the hiring funnel, and keep the candidate record
              visible to the hiring team.
            </p>
          </div>
          <label className="flex w-full max-w-xs items-center gap-2 rounded-md border px-3 py-2 text-sm text-white/60 sm:w-72">
            <Search className="h-4 w-4 shrink-0" />
            <span className="sr-only">Search candidates</span>
            <input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search candidates or roles"
              className="w-full bg-transparent text-white outline-none placeholder:text-white/35"
            />
          </label>
        </div>

        <div className="mt-6 grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
          {PIPELINE.map((status) => {
            const count = filtered.filter((submission) => submission.status === status).length;
            return (
              <div
                key={status}
                className="rounded-lg border px-4 py-3"
                style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
              >
                <div className="flex items-center justify-between text-xs text-white/55">
                  <span>{LABELS[status]}</span>
                  <span className="h-2 w-2 rounded-full" style={{ background: COLORS[status] }} />
                </div>
                <div className="mt-1 text-xl font-semibold">{count}</div>
              </div>
            );
          })}
        </div>

        <div className="mt-6 grid gap-6 lg:grid-cols-[minmax(0,1fr)_340px]">
          <section
            className="min-w-0 overflow-x-auto rounded-lg border"
            style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
            aria-label="Candidate pipeline board"
          >
            {isLoading ? (
              <div className="p-10 text-center text-sm text-white/60">Loading candidates…</div>
            ) : filtered.length === 0 ? (
              <div className="p-10 text-center text-sm text-white/60">
                {search ? 'No candidates match this search.' : 'No applications yet.'}
              </div>
            ) : (
              <div className="grid min-w-[1080px] grid-cols-6 divide-x divide-white/10">
                {PIPELINE.map((status) => {
                  const candidates = filtered.filter((submission) => submission.status === status);
                  return (
                    <div key={status} className="min-h-[420px] p-3">
                      <div className="mb-3 flex items-center justify-between px-1">
                        <span className="text-xs font-semibold uppercase tracking-wide text-white/65">
                          {LABELS[status]}
                        </span>
                        <span className="text-xs text-white/35">{candidates.length}</span>
                      </div>
                      <div className="space-y-2">
                        {candidates.map((submission) => (
                          <CandidateCard
                            key={submission.id}
                            submission={submission}
                            selected={submission.id === selectedId}
                            onSelect={() => setSelectedId(submission.id)}
                          />
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          <CandidateDetail
            submission={selected}
            updatePending={update.isPending}
            onClose={() => setSelectedId(null)}
            onMove={moveCandidate}
          />
        </div>
      </div>
    </>
  );
};

const CandidateCard: React.FC<{
  submission: Submission;
  selected: boolean;
  onSelect: () => void;
}> = ({ submission, selected, onSelect }) => {
  const status = (PIPELINE.includes(submission.status as PipelineStatus)
    ? submission.status
    : 'new') as PipelineStatus;

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onSelect}
      onKeyDown={(event) => {
        if (event.key === 'Enter' || event.key === ' ') onSelect();
      }}
      className="cursor-pointer rounded-md border p-3 transition hover:border-white/30 hover:bg-white/5"
      style={{ borderColor: selected ? COLORS[status] : '#252525', background: '#121212' }}
    >
      <div className="truncate text-sm font-semibold text-white">{submission.candidateName}</div>
      <div className="mt-1 truncate text-xs text-white/55">{submission.careerTitle || 'Open role'}</div>
      <div className="mt-3 truncate text-xs text-white/40">{submission.email}</div>
      <div className="mt-2 text-[11px] text-white/35">
        {new Date(submission.createdAt).toLocaleDateString()}
      </div>
    </div>
  );
};

const CandidateDetail: React.FC<{
  submission: Submission | null;
  updatePending: boolean;
  onClose: () => void;
  onMove: (submission: Submission, status: string) => void;
}> = ({ submission, updatePending, onClose, onMove }) => {
  const [tab, setTab] = useState<'overview' | 'interviews' | 'notes' | 'offers'>('overview');
  const [interviewTitle, setInterviewTitle] = useState('');
  const [interviewType, setInterviewType] = useState('video');
  const [interviewAt, setInterviewAt] = useState('');
  const [interviewerName, setInterviewerName] = useState('');
  const [meetingUrl, setMeetingUrl] = useState('');
  const [noteBody, setNoteBody] = useState('');
  const [offerTitle, setOfferTitle] = useState('');
  const [offerSalary, setOfferSalary] = useState('');
  const [offerCurrency, setOfferCurrency] = useState('USD');
  const [offerStartDate, setOfferStartDate] = useState('');
  const [offerExpiresAt, setOfferExpiresAt] = useState('');
  const [offerTerms, setOfferTerms] = useState('');
  const [tagName, setTagName] = useState('');
  const [selectedPoolId, setSelectedPoolId] = useState('');
  const interviews = useSubmissionInterviews(submission?.id);
  const notes = useSubmissionNotes(submission?.id);
  const offers = useSubmissionOffers(submission?.id);
  const events = useSubmissionEvents(submission?.id);
  const tags = useSubmissionTags(submission?.id);
  const talentPools = useTalentPools();
  const createInterview = useCreateSubmissionInterview();
  const createNote = useCreateSubmissionNote();
  const createOffer = useCreateSubmissionOffer();
  const createTag = useCreateSubmissionTag();
  const deleteTag = useDeleteSubmissionTag();
  const addPoolMember = useAddTalentPoolMember();
  const updateOffer = useUpdateSubmissionOfferStatus();

  if (!submission) {
    return (
      <aside
        className="hidden rounded-lg border p-6 text-sm text-white/50 lg:block"
        style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
      >
        Select a candidate to review their application and update their stage.
      </aside>
    );
  }

  return (
    <aside
      className="rounded-lg border p-5"
      style={{ borderColor: '#1f1f1f', background: '#0f0f0f' }}
      aria-label="Candidate details"
    >
      <div className="flex items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">{submission.candidateName}</h2>
          <p className="mt-1 text-xs text-white/50">{submission.careerTitle || 'Open role'}</p>
        </div>
        <button type="button" onClick={onClose} className="text-white/45 hover:text-white" aria-label="Close candidate details">
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="mt-5 grid grid-cols-4 gap-1 rounded-md border p-1" style={{ borderColor: '#242424' }}>
        {([
          ['overview', 'Overview'],
          ['interviews', 'Interviews'],
          ['notes', 'Notes'],
          ['offers', 'Offers'],
        ] as const).map(([value, label]) => (
          <button
            key={value}
            type="button"
            onClick={() => setTab(value)}
            className="rounded px-1 py-1.5 text-[10px] font-medium"
            style={{ background: tab === value ? '#242424' : 'transparent', color: tab === value ? '#fff' : '#888' }}
          >
            {label}
          </button>
        ))}
      </div>

      {tab === 'overview' && (
        <CandidateOverview
          submission={submission}
          updatePending={updatePending}
          onMove={onMove}
          events={events.data?.events ?? []}
          tags={tags.data?.tags ?? []}
          tagName={tagName}
          setTagName={setTagName}
          onAddTag={() => {
            if (!tagName.trim()) return;
            createTag.mutate(
              { submissionId: submission.id, name: tagName.trim() },
              { onSuccess: () => setTagName('') }
            );
          }}
          tagPending={createTag.isPending}
          onDeleteTag={(tagId) => deleteTag.mutate({ submissionId: submission.id, tagId })}
          talentPools={talentPools.data?.pools ?? []}
          selectedPoolId={selectedPoolId}
          setSelectedPoolId={setSelectedPoolId}
          onAddToPool={() => {
            const poolId = Number(selectedPoolId);
            if (poolId > 0) addPoolMember.mutate({ poolId, submissionId: submission.id });
          }}
          poolPending={addPoolMember.isPending}
        />
      )}

      {tab === 'interviews' && (
        <div className="mt-5 space-y-4">
          <div className="space-y-2 rounded-md border p-3" style={{ borderColor: '#242424' }}>
            <div className="flex items-center gap-2 text-xs font-semibold text-white/65">
              <CalendarDays className="h-3.5 w-3.5" /> Schedule interview
            </div>
            <input value={interviewTitle} onChange={(event) => setInterviewTitle(event.target.value)} placeholder="Interview title" className="w-full rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <div className="grid grid-cols-2 gap-2">
              <select value={interviewType} onChange={(event) => setInterviewType(event.target.value)} className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white">
                <option value="video">Video</option>
                <option value="phone">Phone</option>
                <option value="onsite">On-site</option>
                <option value="panel">Panel</option>
              </select>
              <input type="datetime-local" value={interviewAt} onChange={(event) => setInterviewAt(event.target.value)} className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            </div>
            <input value={interviewerName} onChange={(event) => setInterviewerName(event.target.value)} placeholder="Interviewer" className="w-full rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <input value={meetingUrl} onChange={(event) => setMeetingUrl(event.target.value)} placeholder="Meeting URL (optional)" className="w-full rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <button
              type="button"
              disabled={createInterview.isPending || !interviewTitle || !interviewAt}
              onClick={() => {
                createInterview.mutate(
                  {
                    submissionId: submission.id,
                    interviewType,
                    title: interviewTitle,
                    scheduledAt: new Date(interviewAt).toISOString(),
                    durationMinutes: 60,
                    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC',
                    meetingUrl,
                    interviewerName,
                  },
                  { onSuccess: () => { setInterviewTitle(''); setInterviewAt(''); setMeetingUrl(''); } }
                );
              }}
              className="w-full rounded bg-white px-3 py-1.5 text-xs font-semibold text-black disabled:opacity-40"
            >
              {createInterview.isPending ? 'Scheduling…' : 'Schedule interview'}
            </button>
          </div>
          <div className="space-y-2">
            {(interviews.data?.interviews ?? []).map((interview) => (
              <div key={interview.id} className="rounded-md border p-3 text-xs" style={{ borderColor: '#242424' }}>
                <div className="flex items-center justify-between gap-2 text-white/80"><span className="font-medium">{interview.title}</span><span className="text-[10px] uppercase text-white/40">{interview.status}</span></div>
                <div className="mt-1 text-white/50">{new Date(interview.scheduledAt).toLocaleString()} · {interview.interviewType}</div>
                {interview.interviewerName && <div className="mt-1 text-white/50">With {interview.interviewerName}</div>}
                {interview.meetingUrl && <a href={interview.meetingUrl} target="_blank" rel="noreferrer" className="mt-1 block truncate text-blue-300">Open meeting</a>}
              </div>
            ))}
            {!interviews.isLoading && (interviews.data?.interviews ?? []).length === 0 && <p className="text-xs text-white/40">No interviews scheduled.</p>}
          </div>
        </div>
      )}

      {tab === 'notes' && (
        <div className="mt-5 space-y-4">
          <div className="space-y-2 rounded-md border p-3" style={{ borderColor: '#242424' }}>
            <div className="flex items-center gap-2 text-xs font-semibold text-white/65"><StickyNote className="h-3.5 w-3.5" /> Add recruiter note</div>
            <textarea value={noteBody} onChange={(event) => setNoteBody(event.target.value)} rows={4} placeholder="Keep hiring context with the candidate record…" className="w-full resize-y rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <button type="button" disabled={createNote.isPending || !noteBody.trim()} onClick={() => createNote.mutate({ submissionId: submission.id, body: noteBody }, { onSuccess: () => setNoteBody('') })} className="w-full rounded bg-white px-3 py-1.5 text-xs font-semibold text-black disabled:opacity-40">{createNote.isPending ? 'Saving…' : 'Save note'}</button>
          </div>
          <div className="space-y-2">
            {(notes.data?.notes ?? []).map((note) => (
              <div key={note.id} className="rounded-md border p-3 text-xs text-white/70" style={{ borderColor: '#242424' }}>
                <div className="mb-1 text-[10px] text-white/35">{new Date(note.createdAt).toLocaleString()}</div>
                <p className="whitespace-pre-wrap">{note.body}</p>
              </div>
            ))}
            {!notes.isLoading && (notes.data?.notes ?? []).length === 0 && <p className="text-xs text-white/40">No recruiter notes yet.</p>}
          </div>
        </div>
      )}

      {tab === 'offers' && (
        <div className="mt-5 space-y-4">
          <div className="space-y-2 rounded-md border p-3" style={{ borderColor: '#242424' }}>
            <div className="text-xs font-semibold text-white/65">Create offer</div>
            <input value={offerTitle} onChange={(event) => setOfferTitle(event.target.value)} placeholder="Offer title" className="w-full rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <div className="grid grid-cols-2 gap-2"><input type="number" min="0" value={offerSalary} onChange={(event) => setOfferSalary(event.target.value)} placeholder="Salary" className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white" /><input value={offerCurrency} onChange={(event) => setOfferCurrency(event.target.value.toUpperCase())} maxLength={8} placeholder="Currency" className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white" /></div>
            <div className="grid grid-cols-2 gap-2"><input type="date" value={offerStartDate} onChange={(event) => setOfferStartDate(event.target.value)} className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white" /><input type="date" value={offerExpiresAt} onChange={(event) => setOfferExpiresAt(event.target.value)} className="rounded border bg-black/30 px-2 py-1.5 text-xs text-white" /></div>
            <textarea value={offerTerms} onChange={(event) => setOfferTerms(event.target.value)} rows={3} placeholder="Terms and conditions" className="w-full resize-y rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
            <button type="button" disabled={createOffer.isPending || !offerTitle.trim()} onClick={() => createOffer.mutate({ submissionId: submission.id, title: offerTitle, salary: Number(offerSalary) || 0, salaryCurrency: offerCurrency || 'USD', startDate: offerStartDate, expiresAt: offerExpiresAt, terms: offerTerms }, { onSuccess: () => { setOfferTitle(''); setOfferSalary(''); setOfferStartDate(''); setOfferExpiresAt(''); setOfferTerms(''); } })} className="w-full rounded bg-white px-3 py-1.5 text-xs font-semibold text-black disabled:opacity-40">{createOffer.isPending ? 'Creating…' : 'Create draft offer'}</button>
          </div>
          <div className="space-y-2">
            {(offers.data?.offers ?? []).map((offer) => (
              <div key={offer.id} className="rounded-md border p-3 text-xs" style={{ borderColor: '#242424' }}>
                <div className="flex items-center justify-between gap-2"><span className="font-medium text-white/80">{offer.title}</span><select value={offer.status} onChange={(event) => updateOffer.mutate({ submissionId: submission.id, offerId: offer.id, status: event.target.value })} className="rounded border bg-black/30 px-1 py-1 text-[10px] text-white"><option value="draft">Draft</option><option value="sent">Sent</option><option value="accepted">Accepted</option><option value="declined">Declined</option><option value="expired">Expired</option></select></div>
                <div className="mt-1 text-white/50">{offer.salary ? `${offer.salary.toLocaleString()} ${offer.salaryCurrency}` : 'Compensation not set'}{offer.startDate ? ` · starts ${offer.startDate}` : ''}</div>
                {offer.terms && <p className="mt-2 whitespace-pre-wrap text-white/60">{offer.terms}</p>}
              </div>
            ))}
            {!offers.isLoading && (offers.data?.offers ?? []).length === 0 && <p className="text-xs text-white/40">No offers created.</p>}
          </div>
        </div>
      )}
    </aside>
  );
};

const CandidateOverview: React.FC<{
  submission: Submission;
  updatePending: boolean;
  onMove: (submission: Submission, status: string) => void;
  events: SubmissionEvent[];
  tags: SubmissionTag[];
  tagName: string;
  setTagName: (value: string) => void;
  onAddTag: () => void;
  tagPending: boolean;
  onDeleteTag: (tagId: number) => void;
  talentPools: TalentPool[];
  selectedPoolId: string;
  setSelectedPoolId: (value: string) => void;
  onAddToPool: () => void;
  poolPending: boolean;
}> = ({
  submission,
  updatePending,
  onMove,
  events,
  tags,
  tagName,
  setTagName,
  onAddTag,
  tagPending,
  onDeleteTag,
  talentPools,
  selectedPoolId,
  setSelectedPoolId,
  onAddToPool,
  poolPending,
}) => {
  let answers: Record<string, string | string[]> = {};
  try {
    const parsed = JSON.parse(submission.answersJson || '{}');
    if (parsed && typeof parsed === 'object') answers = parsed as Record<string, string | string[]>;
  } catch {
    answers = {};
  }

  return (
    <>

      <div className="mt-5 space-y-2 text-xs text-white/65">
        <a href={`mailto:${submission.email}`} className="flex items-center gap-2 hover:text-white">
          <Mail className="h-3.5 w-3.5" /> {submission.email}
        </a>
        {submission.phone && (
          <div className="flex items-center gap-2">
            <Phone className="h-3.5 w-3.5" /> {submission.phone}
          </div>
        )}
        {(submission.hasResume ?? Boolean(submission.resumeUrl)) && (
          <button
            type="button"
            className="flex items-center gap-2 hover:text-white"
            onClick={async () => {
              try {
                window.open(await fetchResumeURL(submission.id), '_blank', 'noopener,noreferrer');
              } catch {
                toast.error("Couldn't open resume");
              }
            }}
          >
            <FileText className="h-3.5 w-3.5" /> Resume
          </button>
        )}
      </div>

      <div className="mt-5 space-y-2 rounded-md border p-3" style={{ borderColor: '#242424' }}>
        <div className="flex items-center gap-2 text-xs font-semibold text-white/65">
          <Tag className="h-3.5 w-3.5" /> Candidate tags
        </div>
        <div className="flex flex-wrap gap-1.5">
          {tags.map((tag) => (
            <span key={tag.id} className="inline-flex items-center gap-1 rounded-full bg-white/10 px-2 py-1 text-[11px] text-white/75">
              {tag.name}
              <button type="button" onClick={() => onDeleteTag(tag.id)} className="text-white/45 hover:text-white" aria-label={`Remove ${tag.name} tag`}>
                ×
              </button>
            </span>
          ))}
          {tags.length === 0 && <span className="text-[11px] text-white/35">No tags yet.</span>}
        </div>
        <form
          className="flex gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            onAddTag();
          }}
        >
          <input value={tagName} onChange={(event) => setTagName(event.target.value)} placeholder="Add a tag" maxLength={64} className="min-w-0 flex-1 rounded border bg-black/30 px-2 py-1.5 text-xs text-white" />
          <button type="submit" disabled={tagPending || !tagName.trim()} className="rounded border px-2 py-1.5 text-[11px] text-white/75 disabled:opacity-40">Add</button>
        </form>
      </div>

      <div className="mt-3 space-y-2 rounded-md border p-3" style={{ borderColor: '#242424' }}>
        <div className="flex items-center gap-2 text-xs font-semibold text-white/65">
          <Users className="h-3.5 w-3.5" /> Talent pool
        </div>
        {talentPools.length === 0 ? (
          <p className="text-[11px] text-white/35">Create a talent pool to group this candidate for future roles.</p>
        ) : (
          <div className="flex gap-2">
            <select value={selectedPoolId} onChange={(event) => setSelectedPoolId(event.target.value)} className="min-w-0 flex-1 rounded border bg-black/30 px-2 py-1.5 text-xs text-white">
              <option value="">Select a pool</option>
              {talentPools.map((pool) => <option key={pool.id} value={pool.id}>{pool.name}</option>)}
            </select>
            <button type="button" disabled={poolPending || !selectedPoolId} onClick={onAddToPool} className="rounded border px-2 py-1.5 text-[11px] text-white/75 disabled:opacity-40">Add</button>
          </div>
        )}
      </div>

      <label className="mt-5 block text-xs font-medium text-white/55">
        Stage
        <select
          value={submission.status}
          disabled={updatePending}
          onChange={(event) => onMove(submission, event.target.value)}
          className="mt-2 w-full rounded-md border bg-black/40 px-3 py-2 text-sm text-white outline-none"
          style={{ borderColor: '#2a2a2a' }}
        >
          {PIPELINE.map((status) => (
            <option key={status} value={status}>
              {LABELS[status]}
            </option>
          ))}
        </select>
      </label>

      {events.length > 0 && (
        <div className="mt-5">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-white/45">Activity</h3>
          <div className="mt-2 space-y-2 border-l border-white/10 pl-3">
            {events.slice(0, 8).map((event) => (
              <div key={event.id} className="text-xs text-white/55">
                <span className="text-white/75">{event.eventType === 'status_changed' ? `Moved from ${event.fromStatus || 'unknown'} to ${event.toStatus}` : event.eventType}</span>
                <span className="ml-2 text-[10px] text-white/35">{new Date(event.createdAt).toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {submission.coverLetter && (
        <div className="mt-5">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-white/45">Cover letter</h3>
          <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-white/70">{submission.coverLetter}</p>
        </div>
      )}

      {Object.keys(answers).length > 0 && (
        <div className="mt-5">
          <h3 className="text-xs font-semibold uppercase tracking-wide text-white/45">Application answers</h3>
          <div className="mt-2 space-y-2">
            {Object.entries(answers).map(([key, value]) => (
              <div key={key} className="rounded-md border p-2.5" style={{ borderColor: '#242424' }}>
                <div className="text-[11px] text-white/40">Question {key}</div>
                <div className="mt-1 text-xs text-white/75">
                  {Array.isArray(value) ? value.join(', ') : value || '-'}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </>
  );
};

export default CandidatePipelineAdmin;
