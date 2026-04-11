import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { ArrowLeft, Save } from 'lucide-react';
import { useArticle, useCreateArticle, useUpdateArticle, useArticleCategories } from './useKnowledgeBase';
import { toast } from 'sonner';

const ArticleEditor: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const isEdit = !!id;
  const articleId = Number(id);

  const { data: article, isLoading } = useArticle(articleId);
  const { data: categories } = useArticleCategories();
  const createArticle = useCreateArticle();
  const updateArticle = useUpdateArticle();

  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [category, setCategory] = useState('');

  const categoryList: any[] = categories ?? [];

  useEffect(() => {
    if (article && isEdit) {
      setTitle(article.title || '');
      setBody(article.body || '');
      setCategory(article.category || '');
    }
  }, [article, isEdit]);

  const handleSave = () => {
    if (!title.trim() || !body.trim() || !category.trim()) {
      toast.error('Please fill in all fields');
      return;
    }

    if (isEdit) {
      updateArticle.mutate(
        { id: articleId, data: { title, body, category } },
        {
          onSuccess: () => {
            toast.success('Article updated');
            navigate('/knowledge');
          },
          onError: () => toast.error('Failed to update article'),
        }
      );
    } else {
      createArticle.mutate(
        { title, body, category },
        {
          onSuccess: () => {
            toast.success('Article created');
            navigate('/knowledge');
          },
          onError: () => toast.error('Failed to create article'),
        }
      );
    }
  };

  if (isEdit && isLoading) {
    return (
      <div className="space-y-4">
        <div className="h-8 w-48 rounded animate-pulse" style={{ background: 'var(--hub-hover)' }} />
        <div className="h-64 rounded-xl animate-pulse" style={{ background: 'var(--hub-hover)' }} />
      </div>
    );
  }

  return (
    <div className="max-w-3xl">
      <div className="flex items-center gap-3 mb-6">
        <Link
          to="/knowledge"
          className="h-7 w-7 rounded-lg flex items-center justify-center"
          style={{ color: 'var(--hub-text-muted)', border: '1px solid var(--hub-border)' }}
        >
          <ArrowLeft size={14} />
        </Link>
        <h2 className="text-lg font-semibold flex-1" style={{ color: 'var(--hub-text)' }}>
          {isEdit ? 'Edit Article' : 'New Article'}
        </h2>
        <button
          onClick={handleSave}
          disabled={createArticle.isPending || updateArticle.isPending}
          className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white transition-colors disabled:opacity-40"
          style={{ background: '#4583C4' }}
        >
          <Save size={12} />
          {isEdit ? 'Save changes' : 'Publish'}
        </button>
      </div>

      <div className="space-y-4">
        <div>
          <label className="block text-xs font-medium mb-1.5" style={{ color: 'var(--hub-text-muted)' }}>
            Title
          </label>
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Article title"
            className="w-full h-10 px-3 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </div>

        <div>
          <label className="block text-xs font-medium mb-1.5" style={{ color: 'var(--hub-text-muted)' }}>
            Category
          </label>
          <select
            value={category}
            onChange={(e) => setCategory(e.target.value)}
            className="w-full h-10 px-3 rounded-lg text-sm outline-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          >
            <option value="">Select category</option>
            {categoryList.map((cat: any) => (
              <option key={cat.id || cat.name} value={cat.name}>{cat.name}</option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-xs font-medium mb-1.5" style={{ color: 'var(--hub-text-muted)' }}>
            Body
          </label>
          <textarea
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder="Write your article content..."
            rows={16}
            className="w-full px-3 py-3 rounded-lg text-sm outline-none resize-none"
            style={{
              background: 'var(--hub-input-bg)',
              border: '1px solid var(--hub-border)',
              color: 'var(--hub-text)',
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default ArticleEditor;
