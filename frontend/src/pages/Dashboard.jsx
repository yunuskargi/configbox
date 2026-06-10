import { useEffect, useState } from 'react';
import api from '../api/client';
import { Server, Database, CheckCircle, XCircle, Clock, HardDrive, TrendingUp, Calendar, Activity, PieChart as PieIcon, MapPin, BarChart3, AreaChart as AreaIcon } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, BarChart, Bar, XAxis, YAxis, CartesianGrid, AreaChart, Area } from 'recharts';
import { useTheme } from '../context/ThemeContext';
import { useLang } from '../context/LangContext';

const VENDOR_META = {
  fortigate: { name: 'FortiGate', color: '#f97316', badge: 'bg-orange-100 text-orange-700 dark:bg-orange-900/40 dark:text-orange-300' },
  juniper: { name: 'Juniper', color: '#14b8a6', badge: 'bg-teal-100 text-teal-700 dark:bg-teal-900/40 dark:text-teal-300' },
  cisco: { name: 'Cisco', color: '#6366f1', badge: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/40 dark:text-indigo-300' },
  brocade: { name: 'Brocade', color: '#a855f7', badge: 'bg-purple-100 text-purple-700 dark:bg-purple-900/40 dark:text-purple-300' },
  extreme: { name: 'Extreme', color: '#10b981', badge: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300' },
  paloalto: { name: 'Palo Alto', color: '#ef4444', badge: 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300' },
};
const LOCATION_COLORS = ['#3b82f6', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#06b6d4'];

function StatCard({ label, value, sub, icon: Icon, gradient, progress, progressColor }) {
  return (
    <div className="bg-white dark:bg-gray-900 rounded-2xl p-5 shadow-sm border border-gray-100 dark:border-gray-700 hover:shadow-lg hover:-translate-y-0.5 transition-all duration-200">
      <div className="flex items-start justify-between">
        <div className="min-w-0">
          <p className="text-sm text-gray-500 dark:text-gray-400">{label}</p>
          <p className="text-3xl font-bold text-gray-800 dark:text-gray-100 mt-1 tracking-tight">{value}</p>
          {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-1.5 truncate">{sub}</p>}
        </div>
        <div className={`p-3 rounded-xl bg-gradient-to-br ${gradient} shadow-md shrink-0`}>
          <Icon className="text-white" size={20} />
        </div>
      </div>
      {progress !== undefined && (
        <div className="mt-3 h-1.5 bg-gray-100 dark:bg-gray-800 rounded-full overflow-hidden">
          <div className={`h-full rounded-full transition-all duration-500 ${progressColor}`} style={{ width: `${Math.min(progress, 100)}%` }} />
        </div>
      )}
    </div>
  );
}

function ChartCard({ title, icon: Icon, children, className = '' }) {
  return (
    <div className={`bg-white dark:bg-gray-900 rounded-2xl p-5 shadow-sm border border-gray-100 dark:border-gray-700 ${className}`}>
      <div className="flex items-center gap-2 mb-4">
        <Icon size={15} className="text-gray-400 dark:text-gray-500" />
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wide">{title}</h2>
      </div>
      {children}
    </div>
  );
}

function formatSize(bytes) {
  if (!bytes) return '0 B';
  if (bytes > 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  if (bytes > 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  if (bytes > 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${bytes} B`;
}

export default function Dashboard() {
  const [stats, setStats] = useState(null);
  const [trend, setTrend] = useState([]);
  const [sizeTrend, setSizeTrend] = useState([]);
  const { dark } = useTheme();
  const { lang, t } = useLang();
  const locale = lang === 'tr' ? 'tr-TR' : 'en-US';

  useEffect(() => {
    api.get('/dashboard/stats').then((res) => setStats(res.data));
    api.get('/dashboard/trend?days=30').then((res) => setTrend(res.data));
    api.get('/dashboard/size-trend?days=30').then((res) => setSizeTrend(res.data));
  }, []);

  if (!stats) return (
    <div className="flex items-center justify-center h-64">
      <div className="text-gray-400 flex items-center gap-2"><Activity size={20} className="animate-pulse" /> {t.loading}</div>
    </div>
  );

  const vendorData = Object.entries(stats.vendor_distribution).map(([name, value]) => ({
    name: VENDOR_META[name]?.name || name,
    value,
    fill: VENDOR_META[name]?.color || '#6b7280',
  }));

  const locationData = Object.entries(stats.location_distribution).map(([name, value], i) => ({
    name, value, fill: LOCATION_COLORS[i % LOCATION_COLORS.length],
  }));

  const tooltipStyle = dark
    ? { backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px', color: '#e5e7eb' }
    : { backgroundColor: '#ffffff', border: '1px solid #e5e7eb', borderRadius: '8px' };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t.dash_title}</h1>
        <span className="text-xs text-gray-400">{new Date().toLocaleDateString(locale, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          label={t.dash_total_devices}
          value={stats.total_devices}
          sub={t.dash_active_scheduled(stats.active_devices, stats.scheduled_devices)}
          icon={Server}
          gradient="from-cyan-500 to-blue-500"
        />
        <StatCard
          label={t.dash_total_backups}
          value={stats.total_backups}
          sub={t.dash_total_size(formatSize(stats.total_backup_size))}
          icon={Database}
          gradient="from-emerald-500 to-teal-500"
        />
        <StatCard
          label={t.dash_success_rate}
          value={`%${stats.success_rate}`}
          sub={t.dash_success_failed(stats.successful_backups, stats.failed_backups)}
          icon={TrendingUp}
          gradient={stats.success_rate >= 90 ? 'from-green-500 to-emerald-500' : stats.success_rate >= 70 ? 'from-amber-500 to-orange-500' : 'from-red-500 to-rose-500'}
          progress={stats.success_rate}
          progressColor={stats.success_rate >= 90 ? 'bg-green-500' : stats.success_rate >= 70 ? 'bg-amber-500' : 'bg-red-500'}
        />
        <StatCard
          label={t.dash_today}
          value={stats.today_backups}
          sub={stats.today_failed > 0 ? t.dash_today_failed(stats.today_failed) : t.dash_today_ok}
          icon={Calendar}
          gradient={stats.today_failed > 0 ? 'from-red-500 to-rose-500' : 'from-violet-500 to-purple-500'}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <ChartCard title={t.dash_vendor_dist} icon={PieIcon}>
          {vendorData.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={180}>
                <PieChart>
                  <Pie data={vendorData} cx="50%" cy="50%" innerRadius={45} outerRadius={75} paddingAngle={4} dataKey="value">
                    {vendorData.map((entry, i) => <Cell key={i} fill={entry.fill} stroke="none" />)}
                  </Pie>
                  <Tooltip formatter={(value) => [`${value} ${t.dash_device_unit}`]} contentStyle={tooltipStyle} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex flex-wrap justify-center gap-x-4 gap-y-1.5 mt-2">
                {vendorData.map((v) => (
                  <div key={v.name} className="flex items-center gap-2 text-sm">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: v.fill }} />
                    <span className="text-gray-600 dark:text-gray-300">{v.name} ({v.value})</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_devices}</p>
          )}
        </ChartCard>

        <ChartCard title={t.dash_location_dist} icon={MapPin}>
          {locationData.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={180}>
                <PieChart>
                  <Pie data={locationData} cx="50%" cy="50%" innerRadius={45} outerRadius={75} paddingAngle={4} dataKey="value">
                    {locationData.map((entry, i) => <Cell key={i} fill={entry.fill} stroke="none" />)}
                  </Pie>
                  <Tooltip formatter={(value) => [`${value} ${t.dash_device_unit}`]} contentStyle={tooltipStyle} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex flex-wrap justify-center gap-x-4 gap-y-1.5 mt-2">
                {locationData.map((l) => (
                  <div key={l.name} className="flex items-center gap-2 text-sm">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: l.fill }} />
                    <span className="text-gray-600 dark:text-gray-300">{l.name} ({l.value})</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_locations}</p>
          )}
        </ChartCard>

        <ChartCard title={t.dash_system_status} icon={Activity}>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-xl">
              <div className="flex items-center gap-2.5">
                <div className="p-1.5 rounded-lg bg-cyan-100 dark:bg-cyan-900/40"><HardDrive size={14} className="text-cyan-600 dark:text-cyan-400" /></div>
                <span className="text-sm text-gray-600 dark:text-gray-300">{t.dash_storage}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800 dark:text-gray-100">{formatSize(stats.total_backup_size)}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-xl">
              <div className="flex items-center gap-2.5">
                <div className="p-1.5 rounded-lg bg-violet-100 dark:bg-violet-900/40"><Clock size={14} className="text-violet-600 dark:text-violet-400" /></div>
                <span className="text-sm text-gray-600 dark:text-gray-300">{t.dash_scheduled}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800 dark:text-gray-100">{stats.scheduled_devices} {t.dash_device_unit}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-xl">
              <div className="flex items-center gap-2.5">
                <div className="p-1.5 rounded-lg bg-emerald-100 dark:bg-emerald-900/40"><Server size={14} className="text-emerald-600 dark:text-emerald-400" /></div>
                <span className="text-sm text-gray-600 dark:text-gray-300">{t.dash_active_device}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800 dark:text-gray-100">{stats.active_devices} / {stats.total_devices}</span>
            </div>
            <div className={`flex items-center justify-between p-3 rounded-xl ${stats.success_rate >= 90 ? 'bg-green-50 dark:bg-green-900/20' : stats.success_rate >= 70 ? 'bg-amber-50 dark:bg-amber-900/20' : 'bg-red-50 dark:bg-red-900/20'}`}>
              <div className="flex items-center gap-2.5">
                <div className={`p-1.5 rounded-lg ${stats.success_rate >= 90 ? 'bg-green-100 dark:bg-green-900/40' : stats.success_rate >= 70 ? 'bg-amber-100 dark:bg-amber-900/40' : 'bg-red-100 dark:bg-red-900/40'}`}>
                  <TrendingUp size={14} className={stats.success_rate >= 90 ? 'text-green-600 dark:text-green-400' : stats.success_rate >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-red-600 dark:text-red-400'} />
                </div>
                <span className="text-sm text-gray-600 dark:text-gray-300">{t.dash_success_rate}</span>
              </div>
              <span className={`text-sm font-bold ${stats.success_rate >= 90 ? 'text-green-700 dark:text-green-400' : stats.success_rate >= 70 ? 'text-amber-700 dark:text-amber-400' : 'text-red-700 dark:text-red-400'}`}>%{stats.success_rate}</span>
            </div>
          </div>
        </ChartCard>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ChartCard title={t.dash_backup_trend} icon={BarChart3}>
          {trend.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={trend} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={dark ? '#374151' : '#f1f5f9'} vertical={false} />
                <XAxis dataKey="date" tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(d) => d.slice(5)} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} allowDecimals={false} axisLine={false} tickLine={false} />
                <Tooltip labelFormatter={(d) => new Date(d).toLocaleDateString(locale)} contentStyle={tooltipStyle} cursor={{ fill: dark ? 'rgba(255,255,255,0.04)' : 'rgba(0,0,0,0.03)' }} />
                <Bar dataKey="success" name={t.success} fill="#22c55e" radius={[3, 3, 0, 0]} stackId="a" />
                <Bar dataKey="failed" name={t.failed} fill="#ef4444" radius={[3, 3, 0, 0]} stackId="a" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_data}</p>
          )}
        </ChartCard>

        <ChartCard title={t.dash_storage_trend} icon={AreaIcon}>
          {sizeTrend.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={sizeTrend} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                <defs>
                  <linearGradient id="storageGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#06b6d4" stopOpacity={dark ? 0.5 : 0.35} />
                    <stop offset="100%" stopColor="#06b6d4" stopOpacity={0.02} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke={dark ? '#374151' : '#f1f5f9'} vertical={false} />
                <XAxis dataKey="date" tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(d) => d.slice(5)} axisLine={false} tickLine={false} />
                <YAxis tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(v) => v > 1024 * 1024 ? `${(v / (1024 * 1024)).toFixed(0)}M` : v > 1024 ? `${(v / 1024).toFixed(0)}K` : `${v}B`} axisLine={false} tickLine={false} />
                <Tooltip labelFormatter={(d) => new Date(d).toLocaleDateString(locale)} formatter={(v) => [formatSize(v)]} contentStyle={tooltipStyle} />
                <Area type="monotone" dataKey="cumulative" name={lang === 'tr' ? 'Toplam' : 'Total'} stroke="#06b6d4" fill="url(#storageGradient)" strokeWidth={2.5} />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_data}</p>
          )}
        </ChartCard>
      </div>

      <ChartCard title={t.dash_recent_activity} icon={Database}>
        {stats.recent_activities.length > 0 ? (
          <div className="divide-y divide-gray-50 dark:divide-gray-800">
            {stats.recent_activities.map((a) => (
              <div key={a.id} className="flex items-center justify-between py-3 px-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/60 transition-colors">
                <div className="flex items-center gap-3 min-w-0">
                  <div className={`p-1.5 rounded-full shrink-0 ${a.status === 'success' ? 'bg-green-100 dark:bg-green-900/40' : 'bg-red-100 dark:bg-red-900/40'}`}>
                    {a.status === 'success' ? <CheckCircle size={14} className="text-green-600 dark:text-green-400" /> : <XCircle size={14} className="text-red-600 dark:text-red-400" />}
                  </div>
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate">{a.device_name}</span>
                  <span className={`px-2 py-0.5 rounded-md text-[10px] font-semibold shrink-0 ${VENDOR_META[a.vendor]?.badge || 'bg-gray-100 text-gray-700'}`}>
                    {VENDOR_META[a.vendor]?.name || a.vendor}
                  </span>
                  {a.file_size > 0 && <span className="text-xs text-gray-400 shrink-0">{formatSize(a.file_size)}</span>}
                </div>
                <div className="flex items-center gap-3 shrink-0">
                  <span className={`text-[10px] font-medium px-2 py-0.5 rounded-md ${a.triggered_by === 'manual' ? 'bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'}`}>
                    {a.triggered_by === 'manual' ? t.manual : t.scheduled}
                  </span>
                  <span className="text-xs text-gray-500 dark:text-gray-400">{new Date(a.created_at).toLocaleString(locale)}</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-400 text-sm text-center py-6">{t.dash_no_activity}</p>
        )}
      </ChartCard>
    </div>
  );
}
