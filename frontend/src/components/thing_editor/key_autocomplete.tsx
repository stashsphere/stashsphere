import { useState, useRef, useEffect } from 'react';
import { SchemaCollection } from '../../api/resources';
import { SchemaIcon } from '../shared/svg_icon';
import { Icon } from '../shared';

type Props = {
  collection: SchemaCollection;
  onSelect: (canonicalKey: string) => void;
  onCustom: (key: string) => void;
  selectedKey: string | null;
  isCustom: boolean;
  onClear: () => void;
  usedKeys: Set<string>;
};

type Suggestion = {
  label: string;
  canonicalKey: string;
  isAlias: boolean;
};

export const KeyAutocomplete = ({
  collection,
  onSelect,
  onCustom,
  selectedKey,
  isCustom,
  usedKeys,
}: Props) => {
  const [query, setQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const [highlightIndex, setHighlightIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const suggestions: Suggestion[] = (() => {
    const q = query.toLowerCase().trim();
    if (!q) return [];

    const results: Suggestion[] = [];

    for (const key of Object.keys(collection.schemas)) {
      if (!usedKeys.has(key) && key.toLowerCase().includes(q)) {
        results.push({ label: key, canonicalKey: key, isAlias: false });
      }
    }

    for (const [alias, canonical] of Object.entries(collection.aliases)) {
      if (!usedKeys.has(canonical) && alias.toLowerCase().includes(q)) {
        results.push({ label: alias, canonicalKey: canonical, isAlias: true });
      }
    }

    return results.slice(0, 10);
  })();

  useEffect(() => {
    setHighlightIndex(0);
  }, [query]);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(e.target as Node) &&
        inputRef.current &&
        !inputRef.current.contains(e.target as Node)
      ) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  function handleSelect(suggestion: Suggestion) {
    setQuery('');
    setIsOpen(false);
    onSelect(suggestion.canonicalKey);
  }

  const hasCustomOption = !!query.trim();
  const totalOptions = suggestions.length + (hasCustomOption ? 1 : 0);

  function handleKeyDown(e: React.KeyboardEvent) {
    if (!isOpen || totalOptions === 0) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setHighlightIndex((i) => Math.min(i + 1, totalOptions - 1));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setHighlightIndex((i) => Math.max(i - 1, 0));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (highlightIndex < suggestions.length) {
        handleSelect(suggestions[highlightIndex]);
      } else if (hasCustomOption) {
        onCustom(query.trim());
        setQuery('');
        setIsOpen(false);
      }
    } else if (e.key === 'Escape') {
      setIsOpen(false);
    }
  }

  if (selectedKey || isCustom) {
    return null;
  }

  return (
    <div className="relative">
      <input
        ref={inputRef}
        type="text"
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setIsOpen(true);
        }}
        onFocus={() => {
          if (query.trim()) setIsOpen(true);
        }}
        onKeyDown={handleKeyDown}
        placeholder="Type a property name..."
        className="w-full p-2 text-display border border-secondary shadow-xs focus:border-secondary rounded-sm"
      />
      {isOpen && (suggestions.length > 0 || query.trim()) && (
        <div
          ref={dropdownRef}
          className="absolute z-10 mt-1 w-full bg-white border border-secondary rounded-sm shadow-xs overflow-hidden"
        >
          {suggestions.map((s, i) => (
            <button
              key={`${s.label}-${s.canonicalKey}`}
              className={`w-full flex items-center justify-between px-3 py-2 text-sm text-display cursor-pointer transition-colors ${highlightIndex === i ? 'bg-secondary-800' : 'hover:bg-secondary-900'}`}
              onClick={() => handleSelect(s)}
              onMouseEnter={() => setHighlightIndex(i)}
            >
              <span className="flex items-center gap-2">
                {collection.schemas[s.canonicalKey]?.icons?.[0] && (
                  <SchemaIcon icon={collection.schemas[s.canonicalKey].icons![0]} />
                )}
                <span className="font-medium">{s.label}</span>
              </span>
              {s.isAlias && <span className="text-xs text-display-light">→ {s.canonicalKey}</span>}
            </button>
          ))}
          {query.trim() && (
            <>
              {suggestions.length > 0 && <div className="border-t border-secondary-800" />}
              <button
                className={`w-full flex items-center gap-2 px-3 py-2 text-sm text-display cursor-pointer transition-colors ${highlightIndex === suggestions.length ? 'bg-secondary-800' : 'hover:bg-secondary-900'}`}
                onClick={() => {
                  onCustom(query.trim());
                  setQuery('');
                  setIsOpen(false);
                }}
                onMouseEnter={() => setHighlightIndex(suggestions.length)}
              >
                <Icon icon="mdi--plus" />
                <span>
                  Create custom "<span className="font-semibold">{query.trim()}</span>"
                </span>
              </button>
            </>
          )}
        </div>
      )}
    </div>
  );
};
