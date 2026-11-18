import React, { useState, useEffect, useCallback } from 'react';
import { Navbar } from './components/Navbar';
import { StatsOverview } from './components/StatsOverview';
import { DownloadList } from './components/DownloadList';
import { AddDownloadModal } from './components/AddDownloadModal';
import { SettingsModal } from './components/SettingsModal';
import { DeleteConfirmModal } from './components/DeleteConfirmModal';
import { useWebSocket } from './hooks/useWebSocket';
import { api } from './services/api';

export default function App() {
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [appConfig, setAppConfig] = useState(null);

  // Realtime WebSocket state
  const { isConnected, downloads, setDownloads, stats, setStats } = useWebSocket();

  // Load initial data via REST
  const refreshData = useCallback(async () => {
    try {
      const [items, currentStats, settings] = await Promise.all([
        api.getDownloads(),
        api.getStats(),
        api.getSettings(),
      ]);
      setDownloads(items);
      setStats(currentStats);
      setAppConfig(settings);
    } catch (err) {
      console.error('Failed to refresh data:', err);
    }
  }, [setDownloads, setStats]);

  useEffect(() => {
    refreshData();
  }, [refreshData]);

  // Action handlers
  const handlePause = async (id) => {
    try {
      await api.pauseDownload(id);
      refreshData();
    } catch (err) {
      console.error('Failed to pause download:', err);
    }
  };

  const handleResume = async (id) => {
    try {
      await api.resumeDownload(id);
      refreshData();
    } catch (err) {
      console.error('Failed to resume download:', err);
    }
  };

  const handleDeleteConfirm = async (id, deleteFiles) => {
    try {
      await api.deleteDownload(id, deleteFiles);
      refreshData();
    } catch (err) {
      console.error('Failed to delete download:', err);
    }
  };

  const handleOpenFolder = async (path) => {
    try {
      await api.openFolder(path || appConfig?.download_dir);
    } catch (err) {
      console.error('Failed to open folder:', err);
    }
  };

  return (
    <div className="min-h-screen flex flex-col bg-[#0b0f19] text-slate-100">
      {/* Top Navigation */}
      <Navbar
        isConnected={isConnected}
        stats={stats}
        onOpenAddModal={() => setIsAddOpen(true)}
        onOpenSettings={() => setIsSettingsOpen(true)}
        onOpenFolder={() => handleOpenFolder()}
        onRefresh={refreshData}
      />

      {/* Main Container */}
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* Metric Cards */}
        <StatsOverview stats={stats} />

        {/* Download Queue & Management */}
        <DownloadList
          downloads={downloads}
          onPause={handlePause}
          onResume={handleResume}
          onDelete={(item) => setDeleteTarget(item)}
          onOpenFolder={handleOpenFolder}
          onOpenAddModal={() => setIsAddOpen(true)}
        />
      </main>

      {/* Footer */}
      <footer className="border-t border-slate-900 py-4 text-center text-xs text-slate-600">
        DownForge &bull; High Performance HTTP/S & BitTorrent Manager &bull; Local Service
      </footer>

      {/* Modals */}
      <AddDownloadModal
        isOpen={isAddOpen}
        onClose={() => setIsAddOpen(false)}
        onSuccess={refreshData}
        defaultDir={appConfig?.download_dir}
      />

      <SettingsModal
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        onSaved={(cfg) => setAppConfig(cfg)}
      />

      <DeleteConfirmModal
        isOpen={!!deleteTarget}
        item={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
