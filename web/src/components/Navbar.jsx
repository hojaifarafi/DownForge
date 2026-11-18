import React from 'react';
import {
  Zap,
  Plus,
  Settings as SettingsIcon,
  ArrowDown,
  ArrowUp,
  Folder,
  RefreshCw,
  Wifi,
  WifiOff,
} from 'lucide-react';
import { formatSpeed } from '../services/api';

export function Navbar({
  isConnected,
  stats,
  onOpenAddModal,
  onOpenSettings,
  onOpenFolder,
  onRefresh,
}) {
  return (
    <header className="border-b border-slate-800 bg-slate-900/60 backdrop-blur-md sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-emerald-600 to-teal-400 flex items-center justify-center shadow-lg shadow-emerald-500/20">
            <Zap className="w-6 h-6 text-white fill-white" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xl font-bold tracking-tight text-white">
                Down<span className="text-emerald-400">Forge</span>
              </span>
              <span className="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                v1.0
              </span>
            </div>
            <p className="text-xs text-slate-400">HTTP/S & Torrent Manager</p>
          </div>
        </div>

        {/* Real-time Speeds & Connection Indicator */}
        <div className="hidden md:flex items-center gap-6">
          <div className="flex items-center gap-4 bg-slate-950/60 px-4 py-1.5 rounded-full border border-slate-800">
            {/* Download Speed */}
            <div className="flex items-center gap-1.5 text-xs">
              <ArrowDown className="w-3.5 h-3.5 text-emerald-400" />
              <span className="text-slate-400">Down:</span>
              <span className="font-semibold text-emerald-400 min-w-16">
                {formatSpeed(stats?.total_download_speed || 0)}
              </span>
            </div>

            <div className="w-px h-3.5 bg-slate-800" />

            {/* Upload Speed */}
            <div className="flex items-center gap-1.5 text-xs">
              <ArrowUp className="w-3.5 h-3.5 text-teal-400" />
              <span className="text-slate-400">Up:</span>
              <span className="font-semibold text-teal-400 min-w-16">
                {formatSpeed(stats?.total_upload_speed || 0)}
              </span>
            </div>

            <div className="w-px h-3.5 bg-slate-800" />

            {/* WS Status */}
            <div className="flex items-center gap-1.5 text-xs">
              {isConnected ? (
                <>
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span className="text-slate-400">Live</span>
                </>
              ) : (
                <>
                  <span className="w-2 h-2 rounded-full bg-amber-400" />
                  <span className="text-amber-400">Connecting...</span>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2.5">
          <button
            onClick={onRefresh}
            title="Refresh list"
            className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
          >
            <RefreshCw className="w-4 h-4" />
          </button>

          <button
            onClick={onOpenFolder}
            title="Open Downloads Folder"
            className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
          >
            <Folder className="w-4 h-4" />
          </button>

          <button
            onClick={onOpenSettings}
            title="Settings"
            className="p-2 text-slate-400 hover:text-slate-200 hover:bg-slate-800 rounded-lg transition-colors"
          >
            <SettingsIcon className="w-4 h-4" />
          </button>

          <button
            onClick={onOpenAddModal}
            className="flex items-center gap-2 bg-gradient-to-r from-emerald-500 to-teal-500 hover:from-emerald-600 hover:to-teal-600 text-slate-950 font-semibold px-4 py-2 rounded-lg text-sm shadow-lg shadow-emerald-500/20 transition-all transform hover:-translate-y-0.5 active:translate-y-0"
          >
            <Plus className="w-4 h-4 text-slate-950 stroke-[2.5]" />
            <span>Add Download</span>
          </button>
        </div>
      </div>
    </header>
  );
}
