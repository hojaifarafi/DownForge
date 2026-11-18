import React, { useState } from 'react';
import { X, Link2, FileUp, Globe, Flame, Loader2 } from 'lucide-react';
import { api } from '../services/api';

export function AddDownloadModal({ isOpen, onClose, onSuccess, defaultDir }) {
  const [tab, setTab] = useState('url'); // 'url' | 'torrent'
  const [url, setUrl] = useState('');
  const [name, setName] = useState('');
  const [saveDir, setSaveDir] = useState(defaultDir || '');
  const [torrentFile, setTorrentFile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (!isOpen) return null;

  const isMagnet = url.trim().toLowerCase().startsWith('magnet:');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      if (tab === 'url') {
        if (!url.trim()) {
          throw new Error('Please enter a URL or Magnet link');
        }

        await api.addDownload({
          type: isMagnet ? 'torrent' : 'http',
          url: url.trim(),
          name: name.trim() || undefined,
          save_dir: saveDir.trim() || undefined,
        });
      } else {
        if (!torrentFile) {
          throw new Error('Please select a .torrent file');
        }

        await api.uploadTorrentFile(torrentFile, saveDir.trim() || undefined);
      }

      // Reset form
      setUrl('');
      setName('');
      setTorrentFile(null);
      onSuccess();
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to start download');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800">
          <h2 className="text-base font-semibold text-white">Add New Download</h2>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab switcher */}
        <div className="flex border-b border-slate-800 bg-slate-950/40 p-1">
          <button
            type="button"
            onClick={() => setTab('url')}
            className={`flex-1 py-2 text-xs font-semibold rounded-lg flex items-center justify-center gap-2 transition-all ${
              tab === 'url'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <Link2 className="w-3.5 h-3.5" />
            <span>URL or Magnet</span>
          </button>

          <button
            type="button"
            onClick={() => setTab('torrent')}
            className={`flex-1 py-2 text-xs font-semibold rounded-lg flex items-center justify-center gap-2 transition-all ${
              tab === 'torrent'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            <FileUp className="w-3.5 h-3.5" />
            <span>.torrent File</span>
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs">
              {error}
            </div>
          )}

          {tab === 'url' ? (
            <>
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Download Link or Magnet URI
                </label>
                <div className="relative">
                  <input
                    type="text"
                    required
                    placeholder="https://example.com/file.zip or magnet:?xt=..."
                    value={url}
                    onChange={(e) => setUrl(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2.5 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
                  />
                  {url.trim() && (
                    <div className="absolute right-2.5 top-1/2 -translate-y-1/2">
                      {isMagnet ? (
                        <span className="flex items-center gap-1 text-[10px] bg-purple-500/10 text-purple-400 border border-purple-500/20 px-2 py-0.5 rounded-full font-medium">
                          <Flame className="w-3 h-3" /> Torrent
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded-full font-medium">
                          <Globe className="w-3 h-3" /> HTTP/S
                        </span>
                      )}
                    </div>
                  )}
                </div>
              </div>

              {!isMagnet && (
                <div>
                  <label className="block text-xs font-medium text-slate-300 mb-1.5">
                    Custom File Name <span className="text-slate-500">(Optional)</span>
                  </label>
                  <input
                    type="text"
                    placeholder="e.g. custom_name.iso"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
                  />
                </div>
              )}
            </>
          ) : (
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Select .torrent File
              </label>
              <div className="border-2 border-dashed border-slate-800 hover:border-slate-700 rounded-xl p-6 text-center cursor-pointer bg-slate-950/40 relative">
                <input
                  type="file"
                  accept=".torrent"
                  onChange={(e) => setTorrentFile(e.target.files[0] || null)}
                  className="absolute inset-0 opacity-0 cursor-pointer"
                />
                <FileUp className="w-8 h-8 text-slate-500 mx-auto mb-2" />
                {torrentFile ? (
                  <div>
                    <p className="text-xs font-semibold text-emerald-400 truncate">
                      {torrentFile.name}
                    </p>
                    <p className="text-[10px] text-slate-500 mt-0.5">
                      {(torrentFile.size / 1024).toFixed(1)} KB
                    </p>
                  </div>
                ) : (
                  <div>
                    <p className="text-xs text-slate-300 font-medium">
                      Click to choose or drag & drop a .torrent file
                    </p>
                    <p className="text-[10px] text-slate-500 mt-1">Supports standard BitTorrent files</p>
                  </div>
                )}
              </div>
            </div>
          )}

          <div>
            <label className="block text-xs font-medium text-slate-300 mb-1.5">
              Download Folder <span className="text-slate-500">(Leave blank for default)</span>
            </label>
            <input
              type="text"
              placeholder={defaultDir || 'Default download folder'}
              value={saveDir}
              onChange={(e) => setSaveDir(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
            />
          </div>

          {/* Footer Actions */}
          <div className="flex items-center justify-end gap-3 pt-3 border-t border-slate-800">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex items-center gap-2 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-slate-950 font-semibold px-4 py-2 rounded-lg text-xs shadow-lg shadow-emerald-500/20 transition-all disabled:opacity-50"
            >
              {loading && <Loader2 className="w-3.5 h-3.5 animate-spin" />}
              <span>Start Download</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
