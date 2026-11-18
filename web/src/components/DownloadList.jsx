import React, { useState, useMemo } from 'react';
import { DownloadItem } from './DownloadItem';
import { Search, Inbox, Filter } from 'lucide-react';

export function DownloadList({
  downloads,
  onPause,
  onResume,
  onDelete,
  onOpenFolder,
  onOpenAddModal,
}) {
  const [filter, setFilter] = useState('all');
  const [search, setSearch] = useState('');

  const filteredDownloads = useMemo(() => {
    return downloads.filter((item) => {
      // Status & Type filter
      if (filter === 'downloading' && item.status !== 'downloading') return false;
      if (filter === 'completed' && item.status !== 'completed') return false;
      if (filter === 'paused' && item.status !== 'paused') return false;
      if (filter === 'torrent' && item.type !== 'torrent') return false;
      if (filter === 'http' && item.type !== 'http') return false;

      // Search filter
      if (search.trim()) {
        const query = search.toLowerCase();
        const nameMatches = item.name?.toLowerCase().includes(query);
        const sourceMatches = item.source?.toLowerCase().includes(query);
        if (!nameMatches && !sourceMatches) return false;
      }

      return true;
    });
  }, [downloads, filter, search]);

  const tabs = [
    { id: 'all', label: 'All', count: downloads.length },
    {
      id: 'downloading',
      label: 'Downloading',
      count: downloads.filter((d) => d.status === 'downloading').length,
    },
    {
      id: 'completed',
      label: 'Completed',
      count: downloads.filter((d) => d.status === 'completed').length,
    },
    {
      id: 'paused',
      label: 'Paused',
      count: downloads.filter((d) => d.status === 'paused').length,
    },
    {
      id: 'torrent',
      label: 'Torrents',
      count: downloads.filter((d) => d.type === 'torrent').length,
    },
    {
      id: 'http',
      label: 'HTTP/S',
      count: downloads.filter((d) => d.type === 'http').length,
    },
  ];

  return (
    <div className="space-y-4">
      {/* Controls Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-900/40 p-2 rounded-xl border border-slate-800/80">
        {/* Filter Tabs */}
        <div className="flex items-center gap-1 overflow-x-auto pb-1 sm:pb-0 scrollbar-none">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setFilter(tab.id)}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors flex items-center gap-1.5 ${
                filter === tab.id
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/60'
              }`}
            >
              <span>{tab.label}</span>
              {tab.count > 0 && (
                <span
                  className={`px-1.5 py-0.2 rounded-full text-[10px] ${
                    filter === tab.id
                      ? 'bg-emerald-500/20 text-emerald-300'
                      : 'bg-slate-800 text-slate-500'
                  }`}
                >
                  {tab.count}
                </span>
              )}
            </button>
          ))}
        </div>

        {/* Search Input */}
        <div className="relative min-w-[220px]">
          <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search downloads..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full bg-slate-950/70 border border-slate-800 text-xs rounded-lg pl-9 pr-3 py-1.5 text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>
      </div>

      {/* Downloads List */}
      {filteredDownloads.length > 0 ? (
        <div className="space-y-3">
          {filteredDownloads.map((item) => (
            <DownloadItem
              key={item.id}
              item={item}
              onPause={onPause}
              onResume={onResume}
              onDelete={onDelete}
              onOpenFolder={onOpenFolder}
            />
          ))}
        </div>
      ) : (
        <div className="text-center py-16 px-4 rounded-2xl border border-dashed border-slate-800 bg-slate-900/20">
          <div className="w-12 h-12 rounded-2xl bg-slate-800/60 flex items-center justify-center mx-auto mb-3 text-slate-500">
            <Inbox className="w-6 h-6" />
          </div>
          <h3 className="text-sm font-semibold text-slate-300">
            {search ? 'No matching downloads found' : 'No downloads in this view'}
          </h3>
          <p className="text-xs text-slate-500 max-w-sm mx-auto mt-1 mb-4">
            {search
              ? `No downloads match "${search}". Try clearing your search query.`
              : 'Add an HTTP/HTTPS URL, magnet link, or upload a .torrent file to get started.'}
          </p>
          {!search && (
            <button
              onClick={onOpenAddModal}
              className="inline-flex items-center gap-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/20 font-semibold px-4 py-2 rounded-lg text-xs transition-colors"
            >
              Add First Download
            </button>
          )}
        </div>
      )}
    </div>
  );
}
