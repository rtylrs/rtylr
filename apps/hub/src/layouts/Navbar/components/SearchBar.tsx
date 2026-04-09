import { MagnifyingGlass, X } from "phosphor-react";
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

    const results = searchItems(searchItemsList, searchKeyword, "rtylr");

    const sections: SearchSection[] = [];

    const rtylrItems = results.filter((item) => item.platform === "rtylr");
    if (rtylrItems.length > 0) {
      sections.push({ title: "Rtylr", items: rtylrItems });
    }

    const globalItems = results.filter((item) => item.platform === "global");
    if (globalItems.length > 0) {
      sections.push({ title: "Global", items: globalItems });
    }

    return sections;
  }, [searchKeyword, searchItemsList]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setShowSearchResults(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const keyword = e.target.value;
    setSearchKeyword(keyword);
    setShowSearchResults(!!keyword.trim());
  };

  const handleResultClick = (item: SearchItem) => {
    item.action();
    setSearchKeyword("");
    setShowSearchResults(false);
  };

  const handleSearchFocus = () => {
    if (searchKeyword.trim()) setShowSearchResults(true);
  };

  const clearSearch = () => {
    setSearchKeyword("");
    setShowSearchResults(false);
  };

  return (
    <div className="relative" ref={searchRef}>
      <div className="relative">
        <div className="flex items-center gap-2 bg-black/5 border border-black/8 rounded-lg px-3 py-1.5 w-64">
          <MagnifyingGlass size={14} className="text-muted shrink-0" />
          <input
            type="text"
            placeholder="Search in Rtylr..."
            value={searchKeyword}
            onChange={handleSearchChange}
            onFocus={handleSearchFocus}
            className="flex-1 bg-transparent text-sm text-primary placeholder:text-muted outline-none"
          />
          {searchKeyword && (
            <button onClick={clearSearch} className="text-muted hover:text-primary transition-colors">
              <X size={12} />
            </button>
          )}
        </div>

        {showSearchResults && searchSections.length > 0 && (
          <div className="absolute top-full left-0 mt-1.5 w-72 bg-white border border-black/8 rounded-xl shadow-lg z-50 overflow-hidden py-1">
            {searchSections.map((section) => (
              <div key={section.title}>
                <div className="px-3 py-1.5 text-[10px] font-semibold text-muted uppercase tracking-wider">
                  {section.title}
                </div>
                {section.items.map((item) => (
                  <div
                    key={item.id}
                    className="flex items-center gap-2.5 px-3 py-2 hover:bg-black/4 cursor-pointer"
                    onClick={() => handleResultClick(item)}
                  >
                    <div className="w-6 h-6 rounded-md bg-black/5 flex items-center justify-center text-muted shrink-0">
                      <item.icon size={14} />
                    </div>
                    <div className="min-w-0">
                      <div className="text-sm font-medium text-primary truncate">{item.title}</div>
                      <div className="text-xs text-muted truncate">{item.description}</div>
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
