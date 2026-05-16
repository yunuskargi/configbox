import { useEffect, useState } from 'react';
import api from '../api/client';
import { Shield, Filter } from 'lucide-react';

const actionLabels = {
  login: { label: 'Giriş', color: 'bg-cyan-100 text-cyan-700' },
  login_failed: { label: 'Başarısız Giriş', color: 'bg-red-100 text-red-700' },
  create: { label: 'Oluşturma', color: 'bg-green-100 text-green-700' },
  update: { label: 'Güncelleme', color: 'bg-yellow-100 text-yellow-700' },
  delete: { label: 'Silme', color: 'bg-red-100 text-red-700' },
  backup: { label: 'Backup', color: 'bg-purple-100 text-purple-700' },
};

const resourceLabels = {
  auth: 'Kimlik Doğrulama',
  device: 'Cihaz',
  backup: 'Yedek',
  user: 'Kullanıcı',
  location: 'Lokasyon',
  settings: 'Ayarlar',
};

export default function AuditLog() {
  const [logs, setLogs] = useState([]);
  const [filter, setFilter] = useState({ action: '', resource_type: '' });
  const [showFilters, setShowFilters] = useState(false);

  const load = () => {
    const params = {};
    if (filter.action) params.action = filter.action;
    if (filter.resource_type) params.resource_type = filter.resource_type;
    api.get('/audit', { params }).then((r) => setLogs(r.data));
  };

  useEffect(() => { load(); }, [filter]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Shield className="text-cyan-600" size={24} />
          <h1 className="text-2xl font-bold text-gray-800">Audit Log</h1>
        </div>
        <button onClick={() => setShowFilters(!showFilters)} className="flex items-center gap-2 px-4 py-2 text-sm text-gray-600 border border-gray-200 rounded-lg hover:bg-gray-50">
          <Filter size={16} /> Filtrele
        </button>
      </div>

      {showFilters && (
        <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-4 flex gap-4">
          <select value={filter.action} onChange={(e) => setFilter((f) => ({ ...f, action: e.target.value }))} className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500">
            <option value="">Tüm İşlemler</option>
            {Object.entries(actionLabels).map(([k, v]) => <option key={k} value={k}>{v.label}</option>)}
          </select>
          <select value={filter.resource_type} onChange={(e) => setFilter((f) => ({ ...f, resource_type: e.target.value }))} className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500">
            <option value="">Tüm Kaynaklar</option>
            {Object.entries(resourceLabels).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
          </select>
          <button onClick={() => setFilter({ action: '', resource_type: '' })} className="px-3 py-2 text-sm text-gray-500 hover:text-gray-700">Temizle</button>
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-gray-500 text-left">
            <tr>
              <th className="px-4 py-3 font-medium">Zaman</th>
              <th className="px-4 py-3 font-medium">Kullanıcı</th>
              <th className="px-4 py-3 font-medium">İşlem</th>
              <th className="px-4 py-3 font-medium">Kaynak</th>
              <th className="px-4 py-3 font-medium">Hedef</th>
              <th className="px-4 py-3 font-medium">Detay</th>
              <th className="px-4 py-3 font-medium">IP</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {logs.map((log) => {
              const actionInfo = actionLabels[log.action] || { label: log.action, color: 'bg-gray-100 text-gray-700' };
              return (
                <tr key={log.id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 text-gray-600 text-xs whitespace-nowrap">{new Date(log.created_at).toLocaleString('tr-TR')}</td>
                  <td className="px-4 py-3 font-medium text-gray-800">{log.username}</td>
                  <td className="px-4 py-3">
                    <span className={`px-2 py-0.5 rounded text-xs font-medium ${actionInfo.color}`}>{actionInfo.label}</span>
                  </td>
                  <td className="px-4 py-3 text-gray-600">{resourceLabels[log.resource_type] || log.resource_type}</td>
                  <td className="px-4 py-3 text-gray-800 font-medium">{log.resource_name || '-'}</td>
                  <td className="px-4 py-3 text-gray-500 text-xs max-w-[200px] truncate">{log.detail || '-'}</td>
                  <td className="px-4 py-3 text-gray-400 text-xs font-mono">{log.ip_address || '-'}</td>
                </tr>
              );
            })}
            {logs.length === 0 && (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-gray-400">Henüz kayıt yok</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
