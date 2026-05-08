import { useState, useEffect, type FormEvent } from 'react';
import * as Dialog from '@radix-ui/react-dialog';
import { X } from 'lucide-react';
import { toast } from 'sonner';
import { useMenuStore } from '../../../shared/stores/menuStore';
import { useUserStore } from '../../../shared/stores/userStore';

interface CreateParentCategoryModalProps {
  isOpen: boolean;
  onClose: () => void;
  addOrEdit: 'add' | 'edit';
  categoryData?: {
    _id: string;
    name: string;
    description: string;
    icon: string;
  };
}

const CreateParentCategoryModal: React.FC<CreateParentCategoryModalProps> = ({
  isOpen,
  onClose,
  addOrEdit,
  categoryData,
}) => {
  const [category, setCategory] = useState(categoryData?.name || '');
  const [description, setDescription] = useState(categoryData?.description || '');
  const [iconURL, setIconURL] = useState(categoryData?.icon || '');
  const { addCategory, editCategory } = useMenuStore();
  const { userId: restaurantId } = useUserStore();

  useEffect(() => {
    if (isOpen && categoryData) {
      setCategory(categoryData.name || '');
      setDescription(categoryData.description || '');
      setIconURL(categoryData.icon || '');
    }
  }, [isOpen, categoryData]);

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (category.trim()) {
      const categoryPayload = {
        restaurantId,
        name: category,
        description,
        isParentCategory: true,
        icon: iconURL,
      };

      if (addOrEdit === 'add') {
        addCategory(categoryPayload as any);
      } else if (addOrEdit === 'edit' && categoryData) {
        editCategory(categoryData?._id, categoryPayload as any);
      }

      setCategory('');
      setDescription('');
      setIconURL('');
      onClose();
    } else {
      toast.error('Category name cannot be empty.');
    }
  };

  return (
    <Dialog.Root open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <Dialog.Portal>
        <Dialog.Overlay className="modal-overlay" />
        <Dialog.Content className="category-modal">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Dialog.Title asChild>
              <h2>{addOrEdit === 'add' ? 'Create' : 'Edit'} Parent Category</h2>
            </Dialog.Title>
            <Dialog.Close asChild>
              <button style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 4 }}>
                <X size={18} />
              </button>
            </Dialog.Close>
          </div>
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label>Category Name*</label>
              <input
                type="text"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                required
              />
            </div>
            <div className="form-group">
              <label>Description</label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="e.g. Main Courses, Beverages, etc."
              />
            </div>
            <input
              type="text"
              placeholder="Paste custom icon URL"
              value={iconURL}
              onChange={(e) => setIconURL(e.target.value)}
            />
            <div className="form-actions">
              <button type="submit">Save</button>
              <button type="button" onClick={onClose}>
                Cancel
              </button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
};

export default CreateParentCategoryModal;
