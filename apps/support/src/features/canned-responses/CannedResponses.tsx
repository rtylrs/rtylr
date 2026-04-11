import React, { useState } from 'react';
import { motion } from 'motion/react';
import { MessageSquareText, Search, Plus, Pencil, Trash2, X, Save } from 'lucide-react';
import { useCannedResponses, useCreateCannedResponse, useUpdateCannedResponse, useDeleteCannedResponse } from './useCannedResponses';
import { toast } from 'sonner';

const CannedResponses: React.FC = () => {
  const [search, setSearch] = useState('');
  const { data, isLoading, isError } = useCannedResponses({ search: search || undefined });
  const createMutation = useCreateCannedResponse();
  const updateMutation = useUpdateCannedResponse();
  const deleteMutation = useDeleteCannedResponse();

  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [formTitle, setFormTitle] = useState('');
  const [formBody, setFormBody] = useState('');

  const responses: any[] = data ?? [];

  const resetForm = () => {
    setShowForm(false);
    setEditingId(null);
    setFormTitle('');
    setFormBody('');
  };

  const handleEdit = (item: any) => {
    setEditingId(item.id);
    setFormTitle(item.title);
    setFormBody(item.body);
    setShowForm(true);
  };

  const handleSave = () => {
    if (!formTitle.trim() || !formBody.trim()) {
      toast.error('Please fill in all fields');
      return;
    }

    if (editingId) {
      updateMutation.mutate(
        { id: editingId, data: { title: formTitle, body: formBody } },
        {
          onSuccess: () => { toast.success('Response updated'); resetForm(); },
          onError: () => toast.error('Failed to update response'),
        }
      );
    } else {
      createMutation.mutate(
        { title: formTitle, body: formBody },
        {
          onSuccess: () => { toast.success('Response created'); resetForm(); },
          onError: () => toast.error('Failed to create response'),
        }
      );
    }
  };

  const handleDelete = (id: number) => {
    deleteMutation.mutate(id, {
      onSuccess: () => toast.success('Response deleted'),
      onError: () => toast.error('Failed to delete response'),
    });
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Canned Responses
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Reusable reply templates for common questions
          </p>
        </div>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white transition-colors"
          style={{ background: '#4583C4' }}
        >
          <Plus size={14} />
          New template
        </button>
      </div>

      {showForm && (
        <div
          className="rounded-xl p-4 mb-4"
          style={{ border: '1px solid var(--hub-border)', background: 'var(--hub-card)' }}
        >
          <div className="flex items-center justify-between mb-3">
            <span className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
              {editingId ? 'Edit Template' : 'New Template'}
            </span>
            <button onClick={resetForm} style={{ color: 'var(--hub-text-muted)' }}>
              <X size={14} />
            </button>
          </div>
          <input
            type="text"
            placeholder="Template title"
            value={formTitle}
            onChange={(e) => setFormTitle(e.target.value)}
            className="w-full h-9 px-3 rounded-lg text-sm outline-none mb-3"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
          <textarea
            placeholder="Template body..."
            value={formBody}
            onChange={(e) => setFormBody(e.target.value)}
            rows={4}
            className="w-full px-3 py-2 rounded-lg text-sm outline-none resize-none mb-3"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
          <div className="flex justify-end">
            <button
              onClick={handleSave}
              disabled={createMutation.isPending || updateMutation.isPending}
              className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white transition-colors disabled:opacity-40"
              style={{ background: '#4583C4' }}
            >
              <Save size={12} />
              Save
            </button>
          </div>
        </div>
      )}

      <div className="flex items-center gap-3 mb-4">
        <div
          className="flex items-center gap-2 flex-1 h-9 px-3 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search templates..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
      </div>

      {isLoading && (
        <div className="space-y-2">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="h-20 rounded-lg animate-pulse" style={{ background: 'var(--hub-hover)' }} />
          ))}
        </div>
      )}

      {isError && (
        <div
          className="flex items-center justify-center py-16 rounded-xl text-sm"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          Failed to load canned responses. Please try again.
        </div>
      )}

      {!isLoading && !isError && responses.length === 0 && (
        <div
          className="flex flex-col items-center justify-center py-16 rounded-xl text-sm"
          style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)', color: 'var(--hub-text-muted)' }}
        >
          <MessageSquareText size={32} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p>No canned responses yet</p>
        </div>
      )}

      {!isLoading && !isError && responses.length > 0 && (
        <div className="space-y-2">
          {responses.map((item: any) => (
            <motion.div
              key={item.id}
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              className="rounded-xl p-4 transition-colors"
              style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
            >
              <div className="flex items-start justify-between mb-1">
                <h3 className="text-sm font-medium" style={{ color: 'var(--hub-text)' }}>
                  {item.title}
                </h3>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => handleEdit(item)}
                    className="h-6 w-6 rounded flex items-center justify-center transition-colors"
                    style={{ color: 'var(--hub-text-muted)' }}
                    onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)'; }}
                    onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                  >
                    <Pencil size={11} />
                  </button>
                  <button
                    onClick={() => handleDelete(item.id)}
                    className="h-6 w-6 rounded flex items-center justify-center transition-colors"
                    style={{ color: '#ef4444' }}
                    onMouseEnter={(e) => { (e.currentTarget as HTMLElement).style.background = 'rgba(239,68,68,0.08)'; }}
                    onMouseLeave={(e) => { (e.currentTarget as HTMLElement).style.background = 'transparent'; }}
                  >
                    <Trash2 size={11} />
                  </button>
                </div>
              </div>
              <p className="text-xs leading-relaxed" style={{ color: 'var(--hub-text-muted)' }}>
                {item.body}
              </p>
            </motion.div>
          ))}
        </div>
      )}
    </div>
  );
};

export default CannedResponses;
