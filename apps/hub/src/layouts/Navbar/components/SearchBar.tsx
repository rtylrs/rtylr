import { MagnifyingGlass, X } from "phosphor-react";
import React, { useEffect, useRef, useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { createSearchItems, searchItems, SearchItem } from "../searchItems";
import styles from "../Navbar.module.scss";

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

    const results = searchItems(searchItemsList, searchKeyword, "rtylr");

    const sections: SearchSection[] = [];

    const rtylrItems = results.filter((item) => item.platform === "rtylr");
    if (rtylrItems.length > 0) {
      sections.push({
        title: "Rtylr",
        items: rtylrItems,
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
    <div className={styles.navbarSearch} ref={searchRef}>
      <div className={styles.searchContainer}>
        <div className={styles.searchInputWrapper}>
          <MagnifyingGlass size={14} className={styles.searchIcon} />
          <input
            type="text"
            placeholder="Search in Rtylr..."
            value={searchKeyword}
            onChange={handleSearchChange}
            onFocus={handleSearchFocus}
            className={styles.searchInput}
          />
          {searchKeyword && (
            <button onClick={clearSearch} className={styles.searchClear}>
              <X size={12} />
            </button>
          )}
        </div>

        {showSearchResults && searchSections.length > 0 && (
          <div className={styles.searchResults}>
            {searchSections.map((section) => (
              <div key={section.title} className={styles.searchSection}>
                <div className={styles.searchSectionHeader}>
                  {section.title}
                </div>
                {section.items.map((item) => (
                  <div
                    key={item.id}
                    className={styles.searchResultItem}
                    onClick={() => handleResultClick(item)}
                  >
                    <div className={styles.resultIcon}>
                      <item.icon size={14} />
                    </div>
                    <div className={styles.resultContent}>
                      <div className={styles.resultTitle}>{item.title}</div>
                      <div className={styles.resultDescription}>
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
