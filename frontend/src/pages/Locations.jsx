import { useEffect, useState } from 'react';
import api from '../api/client';
import { Plus, Pencil, Trash2, X, MapPin } from 'lucide-react';

function LocationModal({ location, onClose, onSaved }) {
  const isEdit = !!location?.id;
  const [form, setForm] = useState(location || { name: '', description: '' });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      if (isEdit) {
        await api.put(`/locations/${location.id}`, form);
      } else {
        await api.post('/locations', form);
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
          <h2 className="text-lg font-semibold text-gray-800">{isEdit ? 'Lokasyon Düzenle' : 'Yeni Lokasyon'}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded-lg">{error}</div>}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Lokasyon Adı</label>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Açıklama</label>
            <input value={form.description || ''} onChange={(e) => setForm({ ...form, description: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" placeholder="Opsiyonel" />
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

export default function Locations() {
  const [locations, setLocations] = useState([]);
  const [showModal, setShowModal] = useState(null);
  const [toast, setToast] = useState(null);

  const load = () => api.get('/locations').then((r) => setLocations(r.data));
  useEffect(() => { load(); }, []);

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const handleDelete = async (loc) => {
    if (!confirm(`"${loc.name}" lokasyonunu silmek istediğinize emin misiniz?`)) return;
    try {
      await api.delete(`/locations/${loc.id}`);
      showToast('Lokasyon silindi');
      load();
    } catch (err) {
      showToast(err.response?.data?.detail || 'Hata oluştu', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">Lokasyonlar</h1>
        <button onClick={() => setShowModal({})} className="flex items-center gap-2 bg-cyan-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-cyan-700">
          <Plus size={18} /> Lokasyon Ekle
        </button>
      </div>

      {toast && (
        <div className={`p-3 rounded-lg text-sm ${toast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
          {toast.msg}
        </div>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {locations.map((loc) => (
          <div key={loc.id} className="bg-white rounded-xl p-5 shadow-sm border border-gray-100">
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                <div className="p-2.5 rounded-lg bg-purple-50"><MapPin className="text-purple-600" size={20} /></div>
                <div>
                  <h3 className="font-semibold text-gray-800">{loc.name}</h3>
                  {loc.description && <p className="text-xs text-gray-500 mt-0.5">{loc.description}</p>}
                </div>
              </div>
              <div className="flex gap-1">
                <button onClick={() => setShowModal(loc)} className="p-1.5 rounded hover:bg-gray-100 text-gray-400 hover:text-cyan-600"><Pencil size={15} /></button>
                <button onClick={() => handleDelete(loc)} className="p-1.5 rounded hover:bg-gray-100 text-gray-400 hover:text-red-600"><Trash2 size={15} /></button>
              </div>
            </div>
            <div className="mt-3 pt-3 border-t border-gray-100">
              <span className="text-sm text-gray-500">{loc.device_count} cihaz</span>
            </div>
          </div>
        ))}
        {locations.length === 0 && (
          <div className="col-span-full text-center py-8 text-gray-400">Henüz lokasyon eklenmemiş</div>
        )}
      </div>

      {showModal && <LocationModal location={showModal.id ? showModal : null} onClose={() => setShowModal(null)} onSaved={() => { setShowModal(null); load(); }} />}
    </div>
  );
}
