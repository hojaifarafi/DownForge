const API_BASE = '/api';

export const api = {
  // Downloads
  async getDownloads() {
    const res = await fetch(`${API_BASE}/downloads`);
    if (!res.ok) throw new Error('Failed to fetch downloads');
    return res.json();
  },

  async getDownload(id) {
    const res = await fetch(`${API_BASE}/downloads/${id}`);
    if (!res.ok) throw new Error('Failed to fetch download details');
    return res.json();
  },

  async addDownload(payload) {
    const res = await fetch(`${API_BASE}/downloads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Failed to add download');
    }
    return res.json();
  },

  async uploadTorrentFile(file, saveDir) {
    const formData = new FormData();
    formData.append('torrent', file);
    if (saveDir) {
      formData.append('save_dir', saveDir);
    }

    const res = await fetch(`${API_BASE}/downloads`, {
      method: 'POST',
      body: formData,
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Failed to upload torrent');
    }
    return res.json();
  },

  async pauseDownload(id) {
    const res = await fetch(`${API_BASE}/downloads/${id}/pause`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to pause download');
    return res.json();
  },

  async resumeDownload(id) {
    const res = await fetch(`${API_BASE}/downloads/${id}/resume`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to resume download');
    return res.json();
  },

  async deleteDownload(id, deleteFiles = false) {
    const res = await fetch(`${API_BASE}/downloads/${id}?deleteFiles=${deleteFiles}`, {
      method: 'DELETE',
    });
    if (!res.ok) throw new Error('Failed to delete download');
    return res.json();
  },

  // Stats
  async getStats() {
    const res = await fetch(`${API_BASE}/stats`);
    if (!res.ok) throw new Error('Failed to fetch stats');
    return res.json();
  },

  // Settings
  async getSettings() {
    const res = await fetch(`${API_BASE}/settings`);
    if (!res.ok) throw new Error('Failed to fetch settings');
    return res.json();
  },

  async updateSettings(settings) {
    const res = await fetch(`${API_BASE}/settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(settings),
    });
    if (!res.ok) throw new Error('Failed to update settings');
    return res.json();
  },

  // OS Folder Action
  async openFolder(path) {
    const res = await fetch(`${API_BASE}/open-folder`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path }),
    });
    return res.json();
  },
};

// Formatting utilities
export function formatBytes(bytes, decimals = 2) {
  if (bytes === 0 || !bytes || bytes < 0) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export function formatSpeed(bytesPerSec) {
  if (!bytesPerSec || bytesPerSec <= 0) return '0 B/s';
  return `${formatBytes(bytesPerSec)}/s`;
}

export function formatETA(seconds) {
  if (!seconds || seconds <= 0 || seconds > 86400 * 30) return '--';
  const d = Math.floor(seconds / (3600 * 24));
  const h = Math.floor((seconds % (3600 * 24)) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);

  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}
