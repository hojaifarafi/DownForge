import React, { useState } from 'react';
import { AlertTriangle, Trash2, X, Loader2 } from 'lucide-react';

export function DeleteConfirmModal({ isOpen, item, onClose, onConfirm }) {
  const [deleteFiles, setDeleteFiles] = useState(false);
  const [loading, setLoading] = useState(false);

  if (!isOpen || !item) return null;

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await onConfirm(item.id, deleteFiles);
      onClose();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl">
        <div className="p-6">
          <div className="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-400 border border-rose-500/20 flex items-center justify-center mb-4">
            <AlertTriangle className="w-6 h-6" />
          </div>

          <h3 className="text-base font-semibold text-white mb-1">Remove Download</h3>
          <p className="text-xs text-slate-400 mb-4">
            Are you sure you want to remove{' '}
            <span className="font-semibold text-slate-200">"{item.name}"</span>?
          </p>

          <label className="flex items-center gap-3 p-3 rounded-xl bg-slate-950/60 border border-slate-800 cursor-pointer mb-6">
            <input
              type="checkbox"
              checked={deleteFiles}
              onChange={(e) => setDeleteFiles(e.target.checked)}
              className="w-4 h-4 rounded border-slate-700 text-rose-500 focus:ring-rose-500/20 bg-slate-900"
            />
            <div>
              <span className="text-xs font-medium text-slate-200">
                Delete downloaded files from disk
              </span>
              <p className="text-[10px] text-slate-500">
                Permanently deletes the downloaded data or incomplete parts
              </p>
            </div>
          </label>

          <div className="flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={loading}
              className="px-4 py-2 text-xs font-medium text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={handleConfirm}
              disabled={loading}
              className="flex items-center gap-2 bg-rose-600 hover:bg-rose-700 text-white font-semibold px-4 py-2 rounded-lg text-xs shadow-lg shadow-rose-600/20 transition-all disabled:opacity-50"
            >
              {loading ? (
                <Loader2 className="w-3.5 h-3.5 animate-spin" />
              ) : (
                <Trash2 className="w-3.5 h-3.5" />
              )}
              <span>Remove</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
