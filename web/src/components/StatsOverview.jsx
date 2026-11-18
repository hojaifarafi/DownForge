import React from 'react';
import { DownloadCloud, CheckCircle2, PauseCircle, Gauge } from 'lucide-react';
import { formatSpeed } from '../services/api';

export function StatsOverview({ stats }) {
  const cards = [
    {
      label: 'Active Downloads',
      value: stats?.active_downloads || 0,
      subtext: `${stats?.total_downloads || 0} total tasks`,
      icon: DownloadCloud,
      color: 'emerald',
      textColor: 'text-emerald-400',
      bgColor: 'bg-emerald-500/10',
      borderColor: 'border-emerald-500/20',
    },
    {
      label: 'Completed',
      value: stats?.completed_downloads || 0,
      subtext: 'Finished successfully',
      icon: CheckCircle2,
      color: 'blue',
      textColor: 'text-sky-400',
      bgColor: 'bg-sky-500/10',
      borderColor: 'border-sky-500/20',
    },
    {
      label: 'Paused',
      value: stats?.paused_downloads || 0,
      subtext: `${stats?.failed_downloads || 0} failed / errors`,
      icon: PauseCircle,
      color: 'amber',
      textColor: 'text-amber-400',
      bgColor: 'bg-amber-500/10',
      borderColor: 'border-amber-500/20',
    },
    {
      label: 'Network Speed',
      value: formatSpeed(stats?.total_download_speed || 0),
      subtext: `▲ ${formatSpeed(stats?.total_upload_speed || 0)} upload`,
      icon: Gauge,
      color: 'purple',
      textColor: 'text-purple-400',
      bgColor: 'bg-purple-500/10',
      borderColor: 'border-purple-500/20',
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
      {cards.map((card, i) => {
        const Icon = card.icon;
        return (
          <div
            key={i}
            className="p-4 rounded-xl bg-slate-900/40 border border-slate-800/80 hover:border-slate-700/80 transition-all"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                {card.label}
              </span>
              <div className={`p-2 rounded-lg ${card.bgColor} ${card.textColor} border ${card.borderColor}`}>
                <Icon className="w-4 h-4" />
              </div>
            </div>
            <div className="flex items-baseline gap-2">
              <span className="text-2xl font-bold text-white tracking-tight">{card.value}</span>
            </div>
            <p className="text-xs text-slate-500 mt-1">{card.subtext}</p>
          </div>
        );
      })}
    </div>
  );
}
