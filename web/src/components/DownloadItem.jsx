import React from 'react';
import {
  Globe,
  Flame,
  Play,
  Pause,
  Trash2,
  FolderOpen,
  AlertCircle,
  CheckCircle,
  Users,
  ArrowDown,
  ArrowUp,
  Clock,
} from 'lucide-react';
import { formatBytes, formatSpeed, formatETA } from '../services/api';

export function DownloadItem({
  item,
  onPause,
  onResume,
  onDelete,
  onOpenFolder,
}) {
  const isDownloading = item.status === 'downloading';
  const isCompleted = item.status === 'completed';
  const isPaused = item.status === 'paused';
  const isError = item.status === 'error';

  return (
    <div className="bg-slate-900/60 border border-slate-800/80 hover:border-slate-700 rounded-xl p-4 transition-all group">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-3 mb-3">
        {/* Name and Type */}
        <div className="flex items-start gap-3 min-w-0">
          <div
            className={`p-2.5 rounded-xl shrink-0 ${
              item.type === 'torrent'
                ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
            }`}
          >
            {item.type === 'torrent' ? (
              <Flame className="w-5 h-5" />
            ) : (
              <Globe className="w-5 h-5" />
            )}
          </div>

          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3
                className="text-sm font-semibold text-slate-100 truncate"
                title={item.name}
              >
                {item.name || 'Untitled Download'}
              </h3>
              <span className="text-[10px] font-medium uppercase px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 border border-slate-700 shrink-0">
                {item.type}
              </span>
            </div>
            <p
              className="text-xs text-slate-500 truncate mt-0.5"
              title={item.source}
            >
              {item.source}
            </p>
          </div>
        </div>

        {/* Status Badge & Actions */}
        <div className="flex items-center justify-between md:justify-end gap-3 shrink-0">
          {/* Status Indicator */}
          <div>
            {isDownloading && (
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
                Downloading
              </span>
            )}
            {isCompleted && (
              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-sky-500/10 text-sky-400 border border-sky-500/20">
                <CheckCircle className="w-3.5 h-3.5" />
                Completed
              </span>
            )}
            {isPaused && (
              <span className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
                <Pause className="w-3.5 h-3.5" />
                Paused
              </span>
            )}
            {isError && (
              <span
                className="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-xs font-medium bg-rose-500/10 text-rose-400 border border-rose-500/20"
                title={item.error_message}
              >
                <AlertCircle className="w-3.5 h-3.5" />
                Error
              </span>
            )}
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-1.5">
            {isDownloading && (
              <button
                onClick={() => onPause(item.id)}
                title="Pause"
                className="p-1.5 text-slate-400 hover:text-amber-400 hover:bg-slate-800 rounded-lg transition-colors"
              >
                <Pause className="w-4 h-4" />
              </button>
            )}

            {(isPaused || isError) && (
              <button
                onClick={() => onResume(item.id)}
                title="Resume"
                className="p-1.5 text-slate-400 hover:text-emerald-400 hover:bg-slate-800 rounded-lg transition-colors"
              >
                <Play className="w-4 h-4 fill-current" />
              </button>
            )}

            <button
              onClick={() => onOpenFolder(item.save_path)}
              title="Open Destination Folder"
              className="p-1.5 text-slate-400 hover:text-sky-400 hover:bg-slate-800 rounded-lg transition-colors"
            >
              <FolderOpen className="w-4 h-4" />
            </button>

            <button
              onClick={() => onDelete(item)}
              title="Delete"
              className="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-slate-800 rounded-lg transition-colors"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="w-full bg-slate-950/70 h-2 rounded-full overflow-hidden border border-slate-800/80 mb-2.5">
        <div
          className={`h-full transition-all duration-300 rounded-full ${
            isCompleted
              ? 'bg-gradient-to-r from-sky-500 to-teal-400'
              : isError
              ? 'bg-rose-500'
              : isPaused
              ? 'bg-amber-500/70'
              : 'bg-gradient-to-r from-emerald-500 to-teal-400 shadow-sm shadow-emerald-500/50'
          }`}
          style={{ width: `${Math.min(100, Math.max(0, item.progress || 0))}%` }}
        />
      </div>

      {/* Metrics Footer */}
      <div className="flex flex-wrap items-center justify-between gap-y-1.5 text-xs text-slate-400 font-mono">
        <div className="flex items-center gap-3">
          {/* Downloaded / Total */}
          <span>
            {formatBytes(item.downloaded_bytes)}
            {item.total_bytes > 0 && ` / ${formatBytes(item.total_bytes)}`}
          </span>

          {/* Progress Percentage */}
          <span className="font-semibold text-slate-200">
            {item.progress?.toFixed(1) || '0.0'}%
          </span>

          {/* Torrent Peers */}
          {item.type === 'torrent' && (
            <span className="flex items-center gap-1 text-slate-400 font-sans">
              <Users className="w-3.5 h-3.5 text-purple-400" />
              <span>{item.peers || 0} peers</span>
            </span>
          )}
        </div>

        <div className="flex items-center gap-4">
          {/* Speeds */}
          {isDownloading && (
            <>
              <span className="flex items-center gap-1 text-emerald-400">
                <ArrowDown className="w-3.5 h-3.5" />
                {formatSpeed(item.download_speed)}
              </span>

              {item.type === 'torrent' && item.upload_speed > 0 && (
                <span className="flex items-center gap-1 text-teal-400">
                  <ArrowUp className="w-3.5 h-3.5" />
                  {formatSpeed(item.upload_speed)}
                </span>
              )}

              {/* ETA */}
              {item.eta > 0 && (
                <span className="flex items-center gap-1 text-slate-400 font-sans">
                  <Clock className="w-3.5 h-3.5 text-slate-500" />
                  <span>ETA: {formatETA(item.eta)}</span>
                </span>
              )}
            </>
          )}

          {isError && (
            <span className="text-rose-400 truncate max-w-xs" title={item.error_message}>
              {item.error_message}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}
