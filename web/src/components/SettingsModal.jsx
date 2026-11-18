import React, { useState, useEffect } from 'react';
import { X, Save, Settings as SettingsIcon, Check, Loader2, Folder } from 'lucide-react';
import { api } from '../services/api';

export function SettingsModal({ isOpen, onClose, onSaved }) {
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (isOpen) {
      setLoading(true);
      setError('');
      setSuccess(false);
      api
        .getSettings()
        .then((data) => setConfig(data))
        .catch((err) => setError(err.message))
        .finally(() => setLoading(false));
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setSuccess(false);

    try {
      const updated = await api.updateSettings({
        ...config,
        port: parseInt(config.port, 10),
        max_concurrent: parseInt(config.max_concurrent, 10),
        torrent_port: parseInt(config.torrent_port, 10),
        download_rate_limit: parseInt(config.download_rate_limit, 10) || 0,
        upload_rate_limit: parseInt(config.upload_rate_limit, 10) || 0,
      });
      setConfig(updated);
      setSuccess(true);
      if (onSaved) onSaved(updated);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err.message || 'Failed to update settings');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800">
          <div className="flex items-center gap-2">
            <SettingsIcon className="w-4 h-4 text-emerald-400" />
            <h2 className="text-base font-semibold text-white">Application Settings</h2>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-slate-800 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {loading ? (
          <div className="p-12 text-center text-slate-500">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-2 text-emerald-400" />
            <p className="text-xs">Loading configuration...</p>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {error && (
              <div className="p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs">
                {error}
              </div>
            )}
            {success && (
              <div className="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs flex items-center gap-2">
                <Check className="w-4 h-4" />
                <span>Settings saved successfully!</span>
              </div>
            )}

            {/* Download Directory */}
            <div>
              <label className="block text-xs font-medium text-slate-300 mb-1.5">
                Default Downloads Directory
              </label>
              <div className="flex gap-2">
                <input
                  type="text"
                  required
                  value={config?.download_dir || ''}
                  onChange={(e) =>
                    setConfig({ ...config, download_dir: e.target.value })
                  }
                  className="flex-1 bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
                />
                <button
                  type="button"
                  onClick={() => api.openFolder(config?.download_dir)}
                  title="Open in File Manager"
                  className="p-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition-colors"
                >
                  <Folder className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Port & Max Concurrent */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Web Server Port
                </label>
                <input
                  type="number"
                  min="1024"
                  max="65535"
                  value={config?.port || 8080}
                  onChange={(e) => setConfig({ ...config, port: e.target.value })}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1.5">
                  Max Concurrent Tasks
                </label>
                <input
                  type="number"
                  min="1"
                  max="20"
                  value={config?.max_concurrent || 3}
                  onChange={(e) =>
                    setConfig({ ...config, max_concurrent: e.target.value })
                  }
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:border-emerald-500/60"
                />
              </div>
            </div>

            {/* Torrent Settings */}
            <div className="pt-2 border-t border-slate-800/80">
              <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                BitTorrent Engine Options
              </h4>

              <div className="space-y-3">
                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={config?.enable_dht ?? true}
                    onChange={(e) =>
                      setConfig({ ...config, enable_dht: e.target.checked })
                    }
                    className="w-4 h-4 rounded border-slate-700 text-emerald-500 focus:ring-emerald-500/20 bg-slate-950"
                  />
                  <div>
                    <span className="text-xs font-medium text-slate-200">
                      Enable DHT (Distributed Hash Table)
                    </span>
                    <p className="text-[10px] text-slate-500">
                      Allows finding peers without centralized trackers
                    </p>
                  </div>
                </label>

                <label className="flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={config?.enable_pex ?? true}
                    onChange={(e) =>
                      setConfig({ ...config, enable_pex: e.target.checked })
                    }
                    className="w-4 h-4 rounded border-slate-700 text-emerald-500 focus:ring-emerald-500/20 bg-slate-950"
                  />
                  <div>
                    <span className="text-xs font-medium text-slate-200">
                      Enable Peer Exchange (PEX)
                    </span>
                    <p className="text-[10px] text-slate-500">
                      Discovers more peers through existing connections
                    </p>
                  </div>
                </label>
              </div>
            </div>

            {/* Footer Actions */}
            <div className="flex items-center justify-end gap-3 pt-4 border-t border-slate-800">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
              >
                Close
              </button>
              <button
                type="submit"
                disabled={saving}
                className="flex items-center gap-2 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-slate-950 font-semibold px-4 py-2 rounded-lg text-xs shadow-lg shadow-emerald-500/20 transition-all disabled:opacity-50"
              >
                {saving ? (
                  <Loader2 className="w-3.5 h-3.5 animate-spin" />
                ) : (
                  <Save className="w-3.5 h-3.5" />
                )}
                <span>Save Settings</span>
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
