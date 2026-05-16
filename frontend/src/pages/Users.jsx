import { useEffect, useState } from 'react';
import api from '../api/client';
import { Plus, Pencil, Trash2, X, Shield, ShieldCheck } from 'lucide-react';

function UserModal({ targetUser, onClose, onSaved }) {
  const isEdit = !!targetUser?.id;
  const [form, setForm] = useState(targetUser || { username: '', password: '', role: 'backup_admin' });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      if (isEdit) {
        const payload = { username: form.username, role: form.role };
        if (form.password) payload.password = form.password;
        await api.put(`/users/${targetUser.id}`, payload);
      } else {
        await api.post('/users', form);
      }
      onSaved();
    } catch (err) {
      setError(err.response?.data?.detail || 'Bir hata oluştu');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-md shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-800">{isEdit ? 'Kullanıcı Düzenle' : 'Yeni Kullanıcı'}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded-lg">{error}</div>}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Kullanıcı Adı</label>
            <input value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{isEdit ? 'Yeni Şifre (boş bırakırsan değişmez)' : 'Şifre'}</label>
            <input type="password" value={form.password || ''} onChange={(e) => setForm({ ...form, password: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required={!isEdit} />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Rol</label>
            <select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500">
              <option value="admin">Admin</option>
              <option value="backup_admin">Backup Admin</option>
            </select>
            <p className="text-xs text-gray-400 mt-1">Backup Admin: cihaz ekleme, backup alma ve görüntüleme. Admin: tüm yetkiler.</p>
          </div>
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">İptal</button>
            <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{saving ? 'Kaydediliyor...' : 'Kaydet'}</button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default function Users() {
  const [users, setUsers] = useState([]);
  const [showModal, setShowModal] = useState(null);
  const [toast, setToast] = useState(null);

  const load = () => api.get('/users').then((r) => setUsers(r.data));
  useEffect(() => { load(); }, []);

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const handleDelete = async (u) => {
    if (!confirm(`"${u.username}" kullanıcısını silmek istediğinize emin misiniz?`)) return;
    try {
      await api.delete(`/users/${u.id}`);
      showToast('Kullanıcı silindi');
      load();
    } catch (err) {
      showToast(err.response?.data?.detail || 'Hata oluştu', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">Kullanıcılar</h1>
        <button onClick={() => setShowModal({})} className="flex items-center gap-2 bg-cyan-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-cyan-700">
          <Plus size={18} /> Kullanıcı Ekle
        </button>
      </div>

      {toast && (
        <div className={`p-3 rounded-lg text-sm ${toast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
          {toast.msg}
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-gray-500 text-left">
            <tr>
              <th className="px-4 py-3 font-medium">Kullanıcı</th>
              <th className="px-4 py-3 font-medium">Rol</th>
              <th className="px-4 py-3 font-medium">Oluşturulma</th>
              <th className="px-4 py-3 font-medium text-right">İşlemler</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {users.map((u) => (
              <tr key={u.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-medium text-gray-800">{u.username}</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-medium ${u.role === 'admin' ? 'bg-cyan-100 text-cyan-700' : 'bg-gray-100 text-gray-700'}`}>
                    {u.role === 'admin' ? <ShieldCheck size={12} /> : <Shield size={12} />}
                    {u.role === 'admin' ? 'Admin' : 'Backup Admin'}
                  </span>
                </td>
                <td className="px-4 py-3 text-gray-600 text-xs">{new Date(u.created_at).toLocaleString('tr-TR')}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    <button onClick={() => setShowModal(u)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-cyan-600"><Pencil size={16} /></button>
                    <button onClick={() => handleDelete(u)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600"><Trash2 size={16} /></button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && <UserModal targetUser={showModal.id ? showModal : null} onClose={() => setShowModal(null)} onSaved={() => { setShowModal(null); load(); }} />}
    </div>
  );
}
