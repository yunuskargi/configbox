import { useEffect, useState } from 'react';
import api from '../api/client';
import { Server, Database, CheckCircle, XCircle, Clock, HardDrive, TrendingUp, Calendar, Activity } from 'lucide-react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, BarChart, Bar, XAxis, YAxis, CartesianGrid, AreaChart, Area } from 'recharts';
import { useTheme } from '../context/ThemeContext';
import { useLang } from '../context/LangContext';

const VENDOR_COLORS = { fortigate: '#f97316', juniper: '#14b8a6', cisco: '#6366f1', paloalto: '#ef4444' };
const LOCATION_COLORS = ['#3b82f6', '#8b5cf6', '#ec4899', '#f59e0b', '#10b981', '#06b6d4'];

function StatCard({ label, value, sub, icon: Icon, color, bg }) {
  return (
    <div className="bg-white dark:bg-gray-900 rounded-xl p-5 shadow-sm border border-gray-100 dark:border-gray-700 hover:shadow-md transition-shadow">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm text-gray-500">{label}</p>
          <p className="text-2xl font-bold text-gray-800 dark:text-gray-100 mt-1">{value}</p>
          {sub && <p className="text-xs text-gray-400 mt-1">{sub}</p>}
        </div>
        <div className={`p-2.5 rounded-xl ${bg}`}><Icon className={color} size={22} /></div>
      </div>
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
    name: name === 'fortigate' ? 'FortiGate' : name === 'cisco' ? 'Cisco' : name === 'paloalto' ? 'Palo Alto' : 'Juniper',
    value,
    fill: VENDOR_COLORS[name] || '#6b7280',
  }));

  const locationData = Object.entries(stats.location_distribution).map(([name, value], i) => ({
    name, value, fill: LOCATION_COLORS[i % LOCATION_COLORS.length],
  }));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">{t.dash_title}</h1>
        <span className="text-xs text-gray-400">{new Date().toLocaleDateString(locale, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</span>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label={t.dash_total_devices} value={stats.total_devices} sub={t.dash_active_scheduled(stats.active_devices, stats.scheduled_devices)} icon={Server} color="text-cyan-600" bg="bg-cyan-50" />
        <StatCard label={t.dash_total_backups} value={stats.total_backups} sub={t.dash_total_size(formatSize(stats.total_backup_size))} icon={Database} color="text-emerald-600" bg="bg-emerald-50" />
        <StatCard label={t.dash_success_rate} value={`%${stats.success_rate}`} sub={t.dash_success_failed(stats.successful_backups, stats.failed_backups)} icon={TrendingUp} color="text-green-600" bg="bg-green-50" />
        <StatCard label={t.dash_today} value={stats.today_backups} sub={stats.today_failed > 0 ? t.dash_today_failed(stats.today_failed) : t.dash_today_ok} icon={Calendar} color={stats.today_failed > 0 ? 'text-red-600' : 'text-cyan-600'} bg={stats.today_failed > 0 ? 'bg-red-50' : 'bg-cyan-50'} />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_vendor_dist}</h2>
          {vendorData.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={180}>
                <PieChart>
                  <Pie data={vendorData} cx="50%" cy="50%" innerRadius={45} outerRadius={75} paddingAngle={4} dataKey="value">
                    {vendorData.map((entry, i) => <Cell key={i} fill={entry.fill} />)}
                  </Pie>
                  <Tooltip formatter={(value) => [`${value} ${t.dash_device_unit}`]} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex justify-center gap-4 mt-2">
                {vendorData.map((v) => (
                  <div key={v.name} className="flex items-center gap-2 text-sm">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: v.fill }} />
                    <span className="text-gray-600">{v.name} ({v.value})</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_devices}</p>
          )}
        </div>

        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_location_dist}</h2>
          {locationData.length > 0 ? (
            <>
              <ResponsiveContainer width="100%" height={180}>
                <PieChart>
                  <Pie data={locationData} cx="50%" cy="50%" innerRadius={45} outerRadius={75} paddingAngle={4} dataKey="value">
                    {locationData.map((entry, i) => <Cell key={i} fill={entry.fill} />)}
                  </Pie>
                  <Tooltip formatter={(value) => [`${value} ${t.dash_device_unit}`]} />
                </PieChart>
              </ResponsiveContainer>
              <div className="flex flex-wrap justify-center gap-3 mt-2">
                {locationData.map((l) => (
                  <div key={l.name} className="flex items-center gap-2 text-sm">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: l.fill }} />
                    <span className="text-gray-600">{l.name} ({l.value})</span>
                  </div>
                ))}
              </div>
            </>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_locations}</p>
          )}
        </div>

        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_system_status}</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2">
                <HardDrive size={16} className="text-gray-500" />
                <span className="text-sm text-gray-600">{t.dash_storage}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800">{formatSize(stats.total_backup_size)}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2">
                <Clock size={16} className="text-gray-500" />
                <span className="text-sm text-gray-600">{t.dash_scheduled}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800">{stats.scheduled_devices} {t.dash_device_unit}</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
              <div className="flex items-center gap-2">
                <Server size={16} className="text-gray-500" />
                <span className="text-sm text-gray-600">{t.dash_active_device}</span>
              </div>
              <span className="text-sm font-semibold text-gray-800">{stats.active_devices} / {stats.total_devices}</span>
            </div>
            <div className="flex items-center justify-between p-3 rounded-lg" style={{ backgroundColor: stats.success_rate >= 90 ? (dark ? '#052e16' : '#f0fdf4') : stats.success_rate >= 70 ? (dark ? '#422006' : '#fffbeb') : (dark ? '#2c0b0e' : '#fef2f2') }}>
              <div className="flex items-center gap-2">
                <TrendingUp size={16} className={stats.success_rate >= 90 ? 'text-green-500' : stats.success_rate >= 70 ? 'text-yellow-500' : 'text-red-500'} />
                <span className="text-sm text-gray-600">{t.dash_success_rate}</span>
              </div>
              <span className={`text-sm font-semibold ${stats.success_rate >= 90 ? 'text-green-700' : stats.success_rate >= 70 ? 'text-yellow-700' : 'text-red-700'}`}>%{stats.success_rate}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_backup_trend}</h2>
          {trend.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={trend} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={dark ? '#374151' : '#f1f5f9'} />
                <XAxis dataKey="date" tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(d) => d.slice(5)} />
                <YAxis tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} allowDecimals={false} />
                <Tooltip labelFormatter={(d) => new Date(d).toLocaleDateString(locale)} contentStyle={dark ? { backgroundColor: '#1f2937', border: '1px solid #374151', color: '#e5e7eb' } : undefined} />
                <Bar dataKey="success" name={t.success} fill="#22c55e" radius={[2, 2, 0, 0]} />
                <Bar dataKey="failed" name={t.failed} fill="#ef4444" radius={[2, 2, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_data}</p>
          )}
        </div>

        <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
          <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_storage_trend}</h2>
          {sizeTrend.length > 0 ? (
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={sizeTrend} margin={{ top: 5, right: 5, left: -20, bottom: 5 }}>
                <CartesianGrid strokeDasharray="3 3" stroke={dark ? '#374151' : '#f1f5f9'} />
                <XAxis dataKey="date" tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(d) => d.slice(5)} />
                <YAxis tick={{ fontSize: 10, fill: dark ? '#9ca3af' : '#6b7280' }} tickFormatter={(v) => v > 1024 * 1024 ? `${(v / (1024 * 1024)).toFixed(0)}M` : v > 1024 ? `${(v / 1024).toFixed(0)}K` : `${v}B`} />
                <Tooltip labelFormatter={(d) => new Date(d).toLocaleDateString(locale)} formatter={(v) => [formatSize(v)]} contentStyle={dark ? { backgroundColor: '#1f2937', border: '1px solid #374151', color: '#e5e7eb' } : undefined} />
                <Area type="monotone" dataKey="cumulative" name={lang === 'tr' ? 'Toplam' : 'Total'} stroke="#06b6d4" fill={dark ? '#164e63' : '#cffafe'} strokeWidth={2} />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <p className="text-gray-400 text-sm text-center py-8">{t.dash_no_data}</p>
          )}
        </div>
      </div>

      <div className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
        <h2 className="text-sm font-semibold text-gray-600 uppercase tracking-wide mb-4">{t.dash_recent_activity}</h2>
        {stats.recent_activities.length > 0 ? (
          <div className="space-y-2">
            {stats.recent_activities.map((a) => (
              <div key={a.id} className="flex items-center justify-between py-2.5 px-4 rounded-lg hover:bg-gray-50 transition-colors">
                <div className="flex items-center gap-3">
                  {a.status === 'success' ? <CheckCircle size={16} className="text-green-500" /> : <XCircle size={16} className="text-red-500" />}
                  <span className="text-sm font-medium text-gray-700">{a.device_name}</span>
                  <span className={`px-2 py-0.5 rounded text-[10px] font-medium ${a.vendor === 'fortigate' ? 'bg-orange-100 text-orange-700' : a.vendor === 'cisco' ? 'bg-indigo-100 text-indigo-700' : a.vendor === 'paloalto' ? 'bg-red-100 text-red-700' : 'bg-teal-100 text-teal-700'}`}>
                    {a.vendor === 'fortigate' ? 'FortiGate' : a.vendor === 'cisco' ? 'Cisco' : a.vendor === 'paloalto' ? 'Palo Alto' : 'Juniper'}
                  </span>
                  {a.file_size > 0 && <span className="text-xs text-gray-400">{formatSize(a.file_size)}</span>}
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-gray-400">{a.triggered_by === 'manual' ? t.manual : t.scheduled}</span>
                  <span className="text-xs text-gray-500">{new Date(a.created_at).toLocaleString(locale)}</span>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <p className="text-gray-400 text-sm text-center py-6">{t.dash_no_activity}</p>
        )}
      </div>
    </div>
  );
}
