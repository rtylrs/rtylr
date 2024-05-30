function CategoryDropdown({
  categories,
  selectedCategory,
  onSelectCategory,
  onCreateCategory,
}) {
  return (
    <div>
      <select
        value={selectedCategory}
        onChange={(e) => onSelectCategory(e.target.value)}
      >
        <option value="">All Categories</option>
        {categories.map((cat) => (
          <option key={cat} value={cat}>
            {cat}
          </option>
        ))}
      </select>
      <button onClick={onCreateCategory}>Create Category</button>
    </div>
  );
}

export default CategoryDropdown;
