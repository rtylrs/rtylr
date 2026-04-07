import { Search, X } from "lucide-react";
import React, { useEffect, useRef, useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { createSearchItems, searchItems, SearchItem } from "../searchItems";

interface SearchSection {
  title: string;
  items: SearchItem[];
}

const SearchBar = () => {
  const navigate = useNavigate();
  const [searchKeyword, setSearchKeyword] = useState("");
  const [showSearchResults, setShowSearchResults] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);

  const searchItemsList = useMemo(() => {
    return createSearchItems(navigate);
  }, [navigate]);

  const searchSections = useMemo(() => {
    if (!searchKeyword.trim()) return [];

    const results = searchItems(searchItemsList, searchKeyword, "vbitez");

    const sections: SearchSection[] = [];

    const vbitezItems = results.filter((item) => item.platform === "vbitez");
    if (vbitezItems.length > 0) {
      sections.push({
        title: "Vbitez",
        items: vbitezItems,
      });
    }

    const globalItems = results.filter((item) => item.platform === "global");
    if (globalItems.length > 0) {
      sections.push({
        title: "Global",
        items: globalItems,
      });
    }

    return sections;
  }, [searchKeyword, searchItemsList]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        searchRef.current &&
        !searchRef.current.contains(event.target as Node)
      ) {
        setShowSearchResults(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const keyword = e.target.value;
    setSearchKeyword(keyword);

    if (keyword.trim()) {
      setShowSearchResults(true);
    } else {
      setShowSearchResults(false);
    }
  };

  const handleResultClick = (item: SearchItem) => {
    item.action();
    setSearchKeyword("");
    setShowSearchResults(false);
  };

  const handleSearchFocus = () => {
    if (searchKeyword.trim()) {
      setShowSearchResults(true);
    }
  };

  const clearSearch = () => {
    setSearchKeyword("");
    setShowSearchResults(false);
  };

  return (
    <div className="navbar__search" ref={searchRef}>
      <div className="search__container">
        <div className="search__input-wrapper">
          <Search size={14} className="search__icon" />
          <input
            type="text"
            placeholder="Search in Vbitez..."
            value={searchKeyword}
            onChange={handleSearchChange}
            onFocus={handleSearchFocus}
            className="search__input"
          />
          {searchKeyword && (
            <button onClick={clearSearch} className="search__clear">
              <X size={12} />
            </button>
          )}
        </div>

        {showSearchResults && searchSections.length > 0 && (
          <div className="search__results">
            {searchSections.map((section) => (
              <div key={section.title} className="search__section">
                <div className="search__section-header">{section.title}</div>
                {section.items.map((item) => (
                  <div
                    key={item.id}
                    className="search__result-item"
                    onClick={() => handleResultClick(item)}
                  >
                    <div className="result__icon">
                      <item.icon size={14} />
                    </div>
                    <div className="result__content">
                      <div className="result__title">{item.title}</div>
                      <div className="result__description">
                        {item.description}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default SearchBar;
