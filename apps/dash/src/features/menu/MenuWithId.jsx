import { useEffect, useState } from 'react';
import useAsync from '../../shared/hooks/useAsync';
import api from '../../shared/lib/api';
const menuApi = { getMenuWithMenuId: ({ id }) => api.get(`/menu/${id}`) };
import { useParams } from 'react-router-dom';

function MenuWithId() {
  const [categories, setCategories] = useState([]);
  const [restaurant, setRestaurant] = useState({});
  const [status, setStatus] = useState('idle');
  const { id } = useParams();
  const getMenu = useAsync({
    fn: menuApi.getMenuWithMenuId,
    onSuccess: () => {
      setCategories(getMenu.result.details.menu.items);
      setRestaurant(getMenu.result.details.business);
      setStatus('succeeded');
    },
    onError: () => {
      setStatus('failed');
    },
    callOnMount: false,
  });

  useEffect(() => {
    setStatus('loading');
    getMenu.main({ id });
  }, []);

  const renderMenuItem = (item) => (
    <div
      key={item._id}
      className="flex items-center justify-between p-3 border border-black/10 rounded-lg bg-white"
    >
      <div className="flex items-center gap-3">
        {item?.details?.image && (
          <img
            src={item.details.image}
            alt={item.details.name}
            className="w-12 h-12 rounded-lg object-cover"
          />
        )}
        <div>
          <p className="text-sm font-medium text-primary">{item?.details?.name}</p>
          {item?.details?.description && (
            <p className="text-xs text-muted line-clamp-1">{item.details.description}</p>
          )}
        </div>
      </div>
      {item?.details?.price !== undefined && (
        <span className="text-sm font-semibold text-primary">
          ${(item.details.price / 100).toFixed(2)}
        </span>
      )}
    </div>
  );

  const renderCategorySection = (category) => (
    <div key={category._id} className="border border-black/10 rounded-xl p-4 bg-white mb-3">
      <div className="flex items-center gap-2 mb-3">
        <span className="font-semibold text-primary">{category?.details?.name}</span>
      </div>
      <div className="space-y-2">{category && category.details.menuItems.map(renderMenuItem)}</div>
    </div>
  );

  if (status === 'loading') {
    return <div>Loading...</div>;
  }

  if (status === 'failed') {
    return <div>Error loading menu.</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-primary">{restaurant?.name}'s Menu</h1>
      <div>
        {categories &&
          categories.filter((item) => item.type === 'category').map(renderCategorySection)}
      </div>
    </div>
  );
}

export default MenuWithId;
