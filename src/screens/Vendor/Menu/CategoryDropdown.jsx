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
        {categories &&
          categories?.map((cat, index) => (
            <option key={index} value={cat}>
              {cat.details.name}
            </option>
          ))}
      </select>
      <button onClick={onCreateCategory}>Create Category</button>
    </div>
  );
}

export default CategoryDropdown;
