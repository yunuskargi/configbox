import { useEffect, useState } from 'react';
import api from '../api/client';
import { Download, Trash2, Copy, CheckCircle, XCircle, Lock, X, GitCompare, Plus, Minus, Loader2 } from 'lucide-react';

function DiffModal({ backupA, backupB, onClose }) {
  const [diff, setDiff] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    api.get(`/backups/diff/${backupA.id}/${backupB.id}`)
      .then((r) => setDiff(r.data))
      .catch((err) => setError(err.response?.data?.detail || 'Diff alınamadı'))
      .finally(() => setLoading(false));
  }, [backupA.id, backupB.id]);

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl w-full max-w-4xl shadow-xl max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <div className="flex items-center gap-3">
            <GitCompare size={20} className="text-cyan-600" />
            <div>
              <h2 className="text-lg font-semibold text-gray-800">Config Karşılaştırma</h2>
              <p className="text-xs text-gray-400">
                {backupA.device_name} — {new Date(backupA.created_at).toLocaleString('tr-TR')} vs {new Date(backupB.created_at).toLocaleString('tr-TR')}
              </p>
            </div>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>

        <div className="flex-1 overflow-auto p-4">
          {loading && (
            <div className="flex items-center justify-center py-12">
              <Loader2 size={24} className="animate-spin text-cyan-600" />
              <span className="ml-2 text-gray-500">Karşılaştırılıyor...</span>
            </div>
          )}
          {error && <div className="text-red-600 bg-red-50 p-4 rounded-lg">{error}</div>}
          {diff && (
            <>
              <div className="flex gap-4 mb-4">
                <div className="flex items-center gap-2 px-3 py-1.5 bg-green-50 rounded-lg">
                  <Plus size={14} className="text-green-600" />
                  <span className="text-sm font-medium text-green-700">{diff.stats.added} eklenen</span>
                </div>
                <div className="flex items-center gap-2 px-3 py-1.5 bg-red-50 rounded-lg">
                  <Minus size={14} className="text-red-600" />
                  <span className="text-sm font-medium text-red-700">{diff.stats.removed} silinen</span>
                </div>
              </div>
              {diff.diff.length === 0 ? (
                <div className="text-center py-12 text-gray-400">
                  <CheckCircle size={32} className="mx-auto mb-2 text-green-400" />
                  <p className="text-sm">İki config arasında fark yok</p>
                </div>
              ) : (
                <div className="bg-gray-900 rounded-lg overflow-auto">
                  <pre className="text-xs font-mono leading-relaxed">
                    {diff.diff.map((line, i) => {
                      let bg = '';
                      let color = 'text-gray-300';
                      if (line.startsWith('+++') || line.startsWith('---')) {
                        bg = 'bg-gray-800';
                        color = 'text-gray-400';
                      } else if (line.startsWith('@@')) {
                        bg = 'bg-cyan-900/40';
                        color = 'text-cyan-300';
                      } else if (line.startsWith('+')) {
                        bg = 'bg-green-900/30';
                        color = 'text-green-300';
                      } else if (line.startsWith('-')) {
                        bg = 'bg-red-900/30';
                        color = 'text-red-300';
                      }
                      return (
                        <div key={i} className={`px-4 py-0.5 ${bg} ${color}`}>
                          {line || ' '}
                        </div>
                      );
                    })}
                  </pre>
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function DownloadAuthModal({ backup, onClose }) {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const authRes = await api.post('/backups/authorize-download', { password, backup_id: backup.id });
      const token = authRes.data.download_token;
      const res = await api.get(`/backups/${backup.id}/download?token=${encodeURIComponent(token)}`, { responseType: 'blob' });
      const url = window.URL.createObjectURL(res.data);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${backup.device_name}_${new Date(backup.created_at).toISOString().slice(0, 10)}.conf`;
      a.click();
      window.URL.revokeObjectURL(url);
      onClose();
    } catch (err) {
      setError(err.response?.data?.detail || 'İndirme başarısız');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-sm shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <Lock size={18} className="text-cyan-600" />
            <h2 className="text-lg font-semibold text-gray-800">Kimlik Doğrulama</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>
        <p className="text-sm text-gray-500 mb-4">Backup indirmek için şifrenizi girin.</p>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded-lg">{error}</div>}
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="Şifreniz"
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500"
            autoFocus
            required
          />
          <div className="flex justify-end gap-3">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">İptal</button>
            <button type="submit" disabled={loading} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">
              {loading ? 'Doğrulanıyor...' : 'İndir'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function Backups() {
  const [backups, setBackups] = useState([]);
  const [devices, setDevices] = useState([]);
  const [filters, setFilters] = useState({ device_id: '', vendor: '', status: '' });
  const [copied, setCopied] = useState(null);
  const [downloadTarget, setDownloadTarget] = useState(null);
  const [diffMode, setDiffMode] = useState(false);
  const [diffSelection, setDiffSelection] = useState([]);
  const [showDiff, setShowDiff] = useState(null);

  const load = () => {
    const params = {};
    if (filters.device_id) params.device_id = filters.device_id;
    if (filters.vendor) params.vendor = filters.vendor;
    if (filters.status) params.status = filters.status;
    api.get('/backups', { params }).then((r) => setBackups(r.data));
  };

  useEffect(() => { api.get('/devices').then((r) => setDevices(r.data)); }, []);
  useEffect(() => { load(); }, [filters]);

  const handleDelete = async (id) => {
    if (!confirm('Bu yedeği silmek istediğinize emin misiniz?')) return;
    await api.delete(`/backups/${id}`);
    load();
  };

  const copyPath = (path) => {
    navigator.clipboard.writeText(path);
    setCopied(path);
    setTimeout(() => setCopied(null), 2000);
  };

  const formatSize = (bytes) => {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">Yedekler</h1>
        <div className="flex items-center gap-2">
          {diffMode && diffSelection.length === 2 && (
            <button onClick={() => { setShowDiff(diffSelection); }} className="flex items-center gap-2 bg-cyan-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-cyan-700">
              <GitCompare size={16} /> Karşılaştır
            </button>
          )}
          <button
            onClick={() => { setDiffMode(!diffMode); setDiffSelection([]); }}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium border ${diffMode ? 'bg-cyan-50 border-cyan-300 text-cyan-700' : 'border-gray-300 text-gray-600 hover:bg-gray-50'}`}
          >
            <GitCompare size={16} /> {diffMode ? `Seçili: ${diffSelection.length}/2` : 'Karşılaştır'}
          </button>
        </div>
      </div>

      {diffMode && (
        <div className="bg-cyan-50 border border-cyan-200 rounded-lg px-4 py-3 text-sm text-cyan-700">
          Karşılaştırmak istediğiniz 2 başarılı yedeği seçin. {diffSelection.length === 0 && 'Henüz seçim yapılmadı.'}
          {diffSelection.length === 1 && '1 yedek seçildi, bir tane daha seçin.'}
          {diffSelection.length === 2 && 'Hazır! "Karşılaştır" butonuna tıklayın.'}
        </div>
      )}

      <div className="flex gap-3 flex-wrap">
        <select value={filters.device_id} onChange={(e) => setFilters((f) => ({ ...f, device_id: e.target.value }))} className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500">
          <option value="">Tüm Cihazlar</option>
          {devices.map((d) => <option key={d.id} value={d.id}>{d.name}</option>)}
        </select>
        <select value={filters.vendor} onChange={(e) => setFilters((f) => ({ ...f, vendor: e.target.value }))} className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500">
          <option value="">Tüm Vendor'lar</option>
          <option value="fortigate">FortiGate</option>
          <option value="juniper">Juniper</option>
          <option value="cisco">Cisco</option>
          <option value="paloalto">Palo Alto</option>
        </select>
        <select value={filters.status} onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value }))} className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-cyan-500">
          <option value="">Tüm Durumlar</option>
          <option value="success">Başarılı</option>
          <option value="failed">Başarısız</option>
        </select>
      </div>

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-gray-500 text-left">
            <tr>
              {diffMode && <th className="px-4 py-3 font-medium w-10"></th>}
              <th className="px-4 py-3 font-medium">Cihaz</th>
              <th className="px-4 py-3 font-medium">Vendor</th>
              <th className="px-4 py-3 font-medium">Tarih</th>
              <th className="px-4 py-3 font-medium">Boyut</th>
              <th className="px-4 py-3 font-medium">Durum</th>
              <th className="px-4 py-3 font-medium">Tetikleyen</th>
              <th className="px-4 py-3 font-medium">Dosya Yolu</th>
              <th className="px-4 py-3 font-medium text-right">İşlemler</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {backups.map((b) => {
              const isSelected = diffSelection.some((s) => s.id === b.id);
              return (
              <tr key={b.id} className={`hover:bg-gray-50 ${isSelected ? 'bg-cyan-50' : ''}`}>
                {diffMode && (
                  <td className="px-4 py-3">
                    {b.status === 'success' && (
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => {
                          if (isSelected) {
                            setDiffSelection((s) => s.filter((x) => x.id !== b.id));
                          } else if (diffSelection.length < 2) {
                            setDiffSelection((s) => [...s, b]);
                          }
                        }}
                        className="rounded border-gray-300 text-cyan-600 focus:ring-cyan-500"
                      />
                    )}
                  </td>
                )}
                <td className="px-4 py-3 font-medium text-gray-800">{b.device_name}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${b.vendor === 'fortigate' ? 'bg-orange-100 text-orange-700' : b.vendor === 'cisco' ? 'bg-indigo-100 text-indigo-700' : b.vendor === 'paloalto' ? 'bg-red-100 text-red-700' : 'bg-teal-100 text-teal-700'}`}>
                    {b.vendor === 'fortigate' ? 'FortiGate' : b.vendor === 'cisco' ? 'Cisco' : b.vendor === 'paloalto' ? 'Palo Alto' : 'Juniper'}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-600 text-xs">{new Date(b.created_at).toLocaleString('tr-TR')}</td>
                <td className="px-4 py-3 text-gray-600">{formatSize(b.file_size)}</td>
                <td className="px-4 py-3">
                  {b.status === 'success'
                    ? <span className="flex items-center gap-1 text-green-600"><CheckCircle size={14} />Başarılı</span>
                    : <span className="flex items-center gap-1 text-red-600" title={b.error_message}><XCircle size={14} />Başarısız</span>}
                </td>
                <td className="px-4 py-3 text-gray-500 text-xs">{b.triggered_by === 'manual' ? 'Manuel' : 'Zamanlanmış'}</td>
                <td className="px-4 py-3">
                  <button onClick={() => copyPath(b.file_path)} className="flex items-center gap-1 text-xs text-gray-500 hover:text-cyan-600 font-mono max-w-48 truncate" title={b.file_path}>
                    <Copy size={12} className="shrink-0" />
                    {copied === b.file_path ? 'Kopyalandı!' : b.file_path}
                  </button>
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    {b.status === 'success' && (
                      <button onClick={() => setDownloadTarget(b)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-cyan-600" title="İndir">
                        <Download size={16} />
                      </button>
                    )}
                    <button onClick={() => handleDelete(b.id)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600" title="Sil">
                      <Trash2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
              );
            })}
            {backups.length === 0 && (
              <tr><td colSpan={diffMode ? 9 : 8} className="px-4 py-8 text-center text-gray-400">Henüz yedek yok</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {downloadTarget && <DownloadAuthModal backup={downloadTarget} onClose={() => setDownloadTarget(null)} />}
      {showDiff && <DiffModal backupA={showDiff[0]} backupB={showDiff[1]} onClose={() => { setShowDiff(null); setDiffMode(false); setDiffSelection([]); }} />}
    </div>
  );
}
