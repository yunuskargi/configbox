import { useEffect, useState, useRef } from 'react';
import api from '../api/client';
import { useLang } from '../context/LangContext';
import { Plus, Play, Pencil, Trash2, Wifi, Clock, X, MapPin, Upload, Download, FileSpreadsheet, CheckCircle, XCircle, Loader2, BookOpen } from 'lucide-react';

const vendorDefaults = { fortigate: { port: 443 }, juniper: { port: 22 }, cisco: { port: 22 }, paloalto: { port: 443 } };
const ciscoPlatforms = { ios: 'IOS / IOS-XE', nxos: 'NX-OS (Nexus)', asa: 'ASA (Firewall)' };

function formatCron(cron, t) {
  if (!cron) return null;
  const parts = cron.split(' ');
  if (parts.length !== 5) return cron;
  const [min, hr, , , dow] = parts;
  if (min === '*/15' && hr === '*') return t.dev_schedule_every15;
  if (min === '0' && hr === '*') return t.dev_schedule_hourly;
  if (min === '0' && hr === '*/3') return t.dev_schedule_every3h;
  if (min === '0' && hr === '*/6') return t.dev_schedule_every6h;
  const weekdays = { 0: t.weekday_0, 1: t.weekday_1, 2: t.weekday_2, 3: t.weekday_3, 4: t.weekday_4, 5: t.weekday_5, 6: t.weekday_6 };
  const time = `${hr.padStart(2, '0')}:${min.padStart(2, '0')}`;
  if (dow !== '*') return `${weekdays[dow] || dow} ${time}`;
  return `${t.dev_schedule_daily} ${time}`;
}

function DeviceModal({ device, onClose, onSaved, t }) {
  const isEdit = !!device?.id;
  const [form, setForm] = useState(() => {
    if (device) {
      return { ...device, auth_token: device.has_token ? '********' : '', ssh_password: device.has_ssh_password ? '********' : '', enable_password: device.has_enable_password ? '********' : '' };
    }
    return { name: '', vendor: 'fortigate', ip_address: '', port: 443, location_id: '', vdom: '', auth_token: '', ssh_username: '', ssh_password: '', enable_password: '', platform: 'ios', schedule_cron: '' };
  });
  const [tokenTouched, setTokenTouched] = useState(false);
  const [sshPassTouched, setSshPassTouched] = useState(false);
  const [enablePassTouched, setEnablePassTouched] = useState(false);
  const [locations, setLocations] = useState([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => { api.get('/locations').then((r) => setLocations(r.data)); }, []);

  const set = (k, v) => setForm((f) => ({ ...f, [k]: v }));

  const parseCron = (cron) => {
    if (!cron) return { type: 'none', hour: '02', minute: '00', weekday: '1' };
    const parts = cron.split(' ');
    if (parts.length !== 5) return { type: 'none', hour: '02', minute: '00', weekday: '1' };
    const [min, hr, , , dow] = parts;
    if (min === '*/15' && hr === '*') return { type: 'every15', hour: '02', minute: '00', weekday: '1' };
    if (min === '0' && hr === '*') return { type: 'hourly', hour: '02', minute: '00', weekday: '1' };
    if (min === '0' && hr === '*/3') return { type: 'every3h', hour: '02', minute: '00', weekday: '1' };
    if (min === '0' && hr === '*/6') return { type: 'every6h', hour: '02', minute: '00', weekday: '1' };
    if (dow !== '*') return { type: 'weekly', hour: hr.padStart(2, '0'), minute: min.padStart(2, '0'), weekday: dow };
    return { type: 'daily', hour: hr.padStart(2, '0'), minute: min.padStart(2, '0'), weekday: '1' };
  };

  const [schedule, setSchedule] = useState(() => parseCron(form.schedule_cron));

  const buildCron = (s) => {
    if (s.type === 'none') return '';
    if (s.type === 'every15') return '*/15 * * * *';
    if (s.type === 'hourly') return '0 * * * *';
    if (s.type === 'every3h') return '0 */3 * * *';
    if (s.type === 'every6h') return '0 */6 * * *';
    if (s.type === 'daily') return `${parseInt(s.minute)} ${parseInt(s.hour)} * * *`;
    if (s.type === 'weekly') return `${parseInt(s.minute)} ${parseInt(s.hour)} * * ${s.weekday}`;
    return '';
  };

  const updateSchedule = (key, val) => {
    const next = { ...schedule, [key]: val };
    setSchedule(next);
    set('schedule_cron', buildCron(next));
  };

  const handleVendorChange = (v) => {
    set('vendor', v);
    set('port', vendorDefaults[v]?.port || 443);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      const payload = { ...form, port: Number(form.port), location_id: form.location_id || null };
      if (isEdit) {
        if (!tokenTouched) delete payload.auth_token;
        if (!sshPassTouched) delete payload.ssh_password;
        if (!enablePassTouched) delete payload.enable_password;
        await api.put(`/devices/${device.id}`, payload);
      } else {
        await api.post('/devices', payload);
      }
      onSaved();
    } catch (err) {
      setError(err.response?.data?.detail || t.error_occurred);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-lg shadow-xl max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-800">{isEdit ? t.dev_edit : t.dev_new}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded-lg">{error}</div>}

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_name}</label>
              <input value={form.name} onChange={(e) => set('name', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_vendor}</label>
              <select value={form.vendor} onChange={(e) => handleVendorChange(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" disabled={isEdit}>
                <option value="fortigate">FortiGate</option>
                <option value="juniper">Juniper</option>
                <option value="cisco">Cisco</option>
                <option value="paloalto">Palo Alto</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_ip}</label>
              <input value={form.ip_address} onChange={(e) => set('ip_address', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_port}</label>
              <input type="number" value={form.port} onChange={(e) => set('port', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_location}</label>
            <select value={form.location_id || ''} onChange={(e) => set('location_id', e.target.value ? Number(e.target.value) : null)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500">
              <option value="">{t.dev_no_location} ({t.optional})</option>
              {locations.map((l) => <option key={l.id} value={l.id}>{l.name}</option>)}
            </select>
          </div>

          {(form.vendor === 'fortigate' || form.vendor === 'paloalto') && (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_api_token}</label>
                <input value={form.auth_token || ''} onChange={(e) => { set('auth_token', e.target.value); setTokenTouched(true); }} onFocus={() => { if (!tokenTouched && form.auth_token === '********') { set('auth_token', ''); setTokenTouched(true); } }} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" placeholder={form.vendor === 'paloalto' ? 'PAN-OS XML API key' : 'FortiGate REST API token'} />
              </div>
              {form.vendor === 'fortigate' && (
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_vdom}</label>
                <input value={form.vdom || ''} onChange={(e) => set('vdom', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" placeholder={t.optional} />
              </div>
              )}
            </>
          )}

          {(form.vendor === 'juniper' || form.vendor === 'cisco') && (
            <>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_ssh_username}</label>
                  <input value={form.ssh_username || ''} onChange={(e) => set('ssh_username', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_ssh_password}</label>
                  <input type="password" value={form.ssh_password || ''} onChange={(e) => { set('ssh_password', e.target.value); setSshPassTouched(true); }} onFocus={() => { if (!sshPassTouched && form.ssh_password === '********') { set('ssh_password', ''); setSshPassTouched(true); } }} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
                </div>
              </div>
              {form.vendor === 'cisco' && (
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_platform}</label>
                    <select value={form.platform || 'ios'} onChange={(e) => set('platform', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500">
                      {Object.entries(ciscoPlatforms).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">{t.dev_enable_password}</label>
                    <input type="password" value={form.enable_password || ''} onChange={(e) => { set('enable_password', e.target.value); setEnablePassTouched(true); }} onFocus={() => { if (!enablePassTouched && form.enable_password === '********') { set('enable_password', ''); setEnablePassTouched(true); } }} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" placeholder={t.optional} />
                  </div>
                </div>
              )}
            </>
          )}

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">{t.dev_schedule}</label>
            <div className="space-y-3">
              <select value={schedule.type} onChange={(e) => updateSchedule('type', e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500">
                <option value="none">{t.dev_schedule_none}</option>
                <option value="every15">{t.dev_schedule_every15}</option>
                <option value="hourly">{t.dev_schedule_hourly}</option>
                <option value="every3h">{t.dev_schedule_every3h}</option>
                <option value="every6h">{t.dev_schedule_every6h}</option>
                <option value="daily">{t.dev_schedule_daily}</option>
                <option value="weekly">{t.dev_schedule_weekly}</option>
              </select>
              {(schedule.type === 'daily' || schedule.type === 'weekly') && (
                <div className="flex items-center gap-3">
                  {schedule.type === 'weekly' && (
                    <select value={schedule.weekday} onChange={(e) => updateSchedule('weekday', e.target.value)} className="px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 text-sm">
                      <option value="1">{t.weekday_1}</option>
                      <option value="2">{t.weekday_2}</option>
                      <option value="3">{t.weekday_3}</option>
                      <option value="4">{t.weekday_4}</option>
                      <option value="5">{t.weekday_5}</option>
                      <option value="6">{t.weekday_6}</option>
                      <option value="0">{t.weekday_0}</option>
                    </select>
                  )}
                  <div className="flex items-center gap-1">
                    <span className="text-sm text-gray-500">{t.dev_schedule_time}:</span>
                    <select value={schedule.hour} onChange={(e) => updateSchedule('hour', e.target.value)} className="px-2 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 text-sm">
                      {Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0')).map((h) => <option key={h} value={h}>{h}</option>)}
                    </select>
                    <span className="text-gray-400">:</span>
                    <select value={schedule.minute} onChange={(e) => updateSchedule('minute', e.target.value)} className="px-2 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 text-sm">
                      {['00', '15', '30', '45'].map((m) => <option key={m} value={m}>{m}</option>)}
                    </select>
                  </div>
                </div>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t.cancel}</button>
            <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{saving ? t.saving : t.save}</button>
          </div>
        </form>
      </div>
    </div>
  );
}

function DeleteModal({ device, onClose, onDeleted, t }) {
  const [keepBackups, setKeepBackups] = useState(true);
  const [deleting, setDeleting] = useState(false);

  const handleDelete = async () => {
    setDeleting(true);
    await api.delete(`/devices/${device.id}?keep_backups=${keepBackups}`);
    onDeleted();
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-sm shadow-xl">
        <h2 className="text-lg font-semibold text-gray-800 mb-2">{t.dev_confirm_delete(device.name)}</h2>
        <label className="flex items-center gap-2 text-sm text-gray-700 mb-4">
          <input type="checkbox" checked={keepBackups} onChange={(e) => setKeepBackups(e.target.checked)} className="rounded" />
          {t.dev_delete_keep}
        </label>
        <div className="flex justify-end gap-3">
          <button onClick={onClose} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">{t.cancel}</button>
          <button onClick={handleDelete} disabled={deleting} className="px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 disabled:opacity-50">{deleting ? t.loading : t.dev_delete_remove}</button>
        </div>
      </div>
    </div>
  );
}

function BulkImportModal({ onClose, onImported, t }) {
  const [step, setStep] = useState('upload');
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState('');
  const fileRef = useRef(null);

  const handleDownloadTemplate = async () => {
    const res = await api.get('/devices/bulk/template', { responseType: 'blob' });
    const url = window.URL.createObjectURL(res.data);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'confbox_devices_template.csv';
    a.click();
    window.URL.revokeObjectURL(url);
  };

  const handleFileSelect = async (e) => {
    const f = e.target.files[0];
    if (!f) return;
    setFile(f);
    setLoading(true);
    setError('');
    try {
      const formData = new FormData();
      formData.append('file', f);
      const res = await api.post('/devices/bulk/preview', formData);
      setPreview(res.data);
      setStep('preview');
    } catch (err) {
      setError(err.response?.data?.detail || t.error_occurred);
    } finally {
      setLoading(false);
    }
  };

  const handleImport = async () => {
    setLoading(true);
    setError('');
    try {
      const formData = new FormData();
      formData.append('file', file);
      const res = await api.post('/devices/bulk/import', formData);
      setResult(res.data);
      setStep('done');
    } catch (err) {
      setError(err.response?.data?.detail || t.error_occurred);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 w-full max-w-3xl shadow-xl max-h-[90vh] flex flex-col">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <FileSpreadsheet size={20} className="text-cyan-600" />
            <h2 className="text-lg font-semibold text-gray-800">{t.bulk_title}</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>

        {error && <div className="text-sm text-red-600 bg-red-50 p-3 rounded-lg mb-4">{error}</div>}

        {step === 'upload' && (
          <div className="space-y-4">
            <div className="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center hover:border-cyan-400 transition-colors">
              <Upload size={32} className="mx-auto text-gray-400 mb-3" />
              <p className="text-sm text-gray-600 mb-2">{t.bulk_select_file}</p>
              <input ref={fileRef} type="file" accept=".csv" onChange={handleFileSelect} className="hidden" />
              <button onClick={() => fileRef.current?.click()} disabled={loading} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">
                {loading ? <span className="flex items-center gap-2"><Loader2 size={14} className="animate-spin" />{t.loading}</span> : t.bulk_upload}
              </button>
            </div>
            <div className="flex items-center justify-between bg-gray-50 rounded-lg p-4">
              <div>
                <p className="text-sm font-medium text-gray-700">{t.bulk_download_template}</p>
              </div>
              <button onClick={handleDownloadTemplate} className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm text-gray-600 hover:bg-white">
                <Download size={14} /> {t.bulk_download_template}
              </button>
            </div>
          </div>
        )}

        {step === 'preview' && preview && (
          <div className="flex-1 overflow-auto space-y-4">
            <div className="flex gap-4">
              <div className="flex items-center gap-2 px-3 py-1.5 bg-gray-100 rounded-lg">
                <span className="text-sm text-gray-600">{t.bulk_step_preview}: <strong>{preview.total}</strong></span>
              </div>
              <div className="flex items-center gap-2 px-3 py-1.5 bg-green-50 rounded-lg">
                <CheckCircle size={14} className="text-green-600" />
                <span className="text-sm text-green-700">{t.bulk_valid}: <strong>{preview.valid}</strong></span>
              </div>
              {preview.invalid > 0 && (
                <div className="flex items-center gap-2 px-3 py-1.5 bg-red-50 rounded-lg">
                  <XCircle size={14} className="text-red-600" />
                  <span className="text-sm text-red-700">{t.bulk_invalid}: <strong>{preview.invalid}</strong></span>
                </div>
              )}
            </div>

            <div className="border border-gray-200 rounded-lg overflow-auto max-h-80">
              <table className="w-full text-xs">
                <thead className="bg-gray-50 text-gray-500 sticky top-0">
                  <tr>
                    <th className="px-3 py-2 font-medium text-left">#</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_name}</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_vendor}</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_ip}</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_port}</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_location}</th>
                    <th className="px-3 py-2 font-medium text-left">{t.dev_status}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {preview.rows.map((r) => (
                    <tr key={r.row} className={r.valid ? '' : 'bg-red-50'}>
                      <td className="px-3 py-2 text-gray-400">{r.row}</td>
                      <td className="px-3 py-2 font-medium text-gray-800">{r.name}</td>
                      <td className="px-3 py-2">
                        <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${r.vendor === 'fortigate' ? 'bg-orange-100 text-orange-700' : r.vendor === 'cisco' ? 'bg-indigo-100 text-indigo-700' : r.vendor === 'paloalto' ? 'bg-red-100 text-red-700' : 'bg-teal-100 text-teal-700'}`}>
                          {r.vendor === 'fortigate' ? 'FortiGate' : r.vendor === 'cisco' ? 'Cisco' : r.vendor === 'paloalto' ? 'Palo Alto' : 'Juniper'}
                        </span>
                      </td>
                      <td className="px-3 py-2 text-gray-600">{r.ip_address}</td>
                      <td className="px-3 py-2 text-gray-600">{r.port}</td>
                      <td className="px-3 py-2 text-gray-600">{r.location || '-'}</td>
                      <td className="px-3 py-2">
                        {r.valid ? (
                          <span className="flex items-center gap-1 text-green-600"><CheckCircle size={12} />{t.bulk_valid}</span>
                        ) : (
                          <span className="text-red-600" title={r.errors.join(', ')}>{r.errors.join(', ')}</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="flex justify-between items-center pt-2">
              <button onClick={() => { setStep('upload'); setFile(null); setPreview(null); }} className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800">
                {t.cancel}
              </button>
              <button onClick={handleImport} disabled={loading || preview.valid === 0} className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">
                {loading ? <Loader2 size={14} className="animate-spin" /> : <Upload size={14} />}
                {t.bulk_import} ({preview.valid})
              </button>
            </div>
          </div>
        )}

        {step === 'done' && result && (
          <div className="text-center py-8 space-y-4">
            <CheckCircle size={48} className="mx-auto text-green-500" />
            <div>
              <p className="text-lg font-semibold text-gray-800">{t.bulk_success(result.created)}</p>
              {result.skipped > 0 && <p className="text-sm text-gray-500">{result.skipped} {t.bulk_row}</p>}
            </div>
            <button onClick={() => { onImported(); onClose(); }} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700">
              {t.save}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

function VendorGuide({ onClose, t }) {
  const [active, setActive] = useState('fortigate');
  const vendors = [
    { id: 'fortigate', name: 'FortiGate', color: 'text-red-600 bg-red-50 border-red-200' },
    { id: 'juniper', name: 'Juniper', color: 'text-green-600 bg-green-50 border-green-200' },
    { id: 'cisco', name: 'Cisco', color: 'text-blue-600 bg-blue-50 border-blue-200' },
    { id: 'paloalto', name: 'Palo Alto', color: 'text-orange-600 bg-orange-50 border-orange-200' },
  ];
  const guides = t.guide_vendors;

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-2xl shadow-xl w-full max-w-3xl max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
          <div className="flex items-center gap-2">
            <BookOpen size={20} className="text-cyan-600" />
            <h2 className="text-lg font-semibold text-gray-800">{t.guide_title}</h2>
          </div>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
        </div>

        <div className="flex gap-2 px-6 py-3 border-b border-gray-100 overflow-x-auto">
          {vendors.map((v) => (
            <button key={v.id} onClick={() => setActive(v.id)}
              className={`px-3 py-1.5 rounded-lg text-sm font-medium border transition-all whitespace-nowrap ${active === v.id ? v.color : 'text-gray-500 bg-gray-50 border-gray-200 hover:bg-gray-100'}`}>
              {v.name}
            </button>
          ))}
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4">
          {guides[active] && (
            <div className="space-y-4">
              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-1">{t.guide_protocol}</h3>
                <span className="inline-block px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-700">{guides[active].protocol}</span>
              </div>

              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-1">{t.guide_port}</h3>
                <span className="text-sm text-gray-600">{guides[active].port}</span>
              </div>

              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-2">{t.guide_requirements}</h3>
                <ul className="space-y-1.5">
                  {guides[active].requirements.map((req, i) => (
                    <li key={i} className="flex items-start gap-2 text-sm text-gray-600">
                      <CheckCircle size={14} className="text-green-500 mt-0.5 shrink-0" />
                      {req}
                    </li>
                  ))}
                </ul>
              </div>

              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-2">{t.guide_setup_steps}</h3>
                <ol className="space-y-2">
                  {guides[active].steps.map((step, i) => (
                    <li key={i} className="flex items-start gap-3 text-sm text-gray-600">
                      <span className="flex items-center justify-center w-5 h-5 rounded-full bg-cyan-100 text-cyan-700 text-xs font-bold shrink-0 mt-0.5">{i + 1}</span>
                      {step}
                    </li>
                  ))}
                </ol>
              </div>

              {guides[active].commands && (
                <div>
                  <h3 className="text-sm font-semibold text-gray-700 mb-2">{t.guide_device_commands}</h3>
                  <pre className="bg-gray-900 text-green-400 rounded-lg p-4 text-xs font-mono overflow-x-auto whitespace-pre-wrap">{guides[active].commands}</pre>
                </div>
              )}

              <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
                <p className="text-xs text-amber-800"><strong>{t.guide_note}:</strong> {guides[active].note}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function Devices() {
  const { lang, t } = useLang();
  const locale = lang === 'tr' ? 'tr-TR' : 'en-US';
  const [devices, setDevices] = useState([]);
  const [showModal, setShowModal] = useState(null);
  const [deleteTarget, setDeleteTarget] = useState(null);
  const [backingUp, setBackingUp] = useState(null);
  const [testing, setTesting] = useState(null);
  const [toast, setToast] = useState(null);
  const [errorDetail, setErrorDetail] = useState(null);
  const [showBulkImport, setShowBulkImport] = useState(false);
  const [showGuide, setShowGuide] = useState(false);
  const [vendorFilter, setVendorFilter] = useState('all');

  const load = () => api.get('/devices').then((r) => setDevices(r.data));
  useEffect(() => { load(); }, []);

  const showToast = (msg, type = 'success', detail = null) => {
    setToast({ msg, type, detail });
    if (type === 'success') setTimeout(() => setToast(null), 3000);
  };

  const handleBackup = async (id) => {
    const device = devices.find((d) => d.id === id);
    setBackingUp(id);
    try {
      const res = await api.post(`/devices/${id}/backup`);
      if (res.data.status === 'success') {
        showToast(t.success, 'success');
      } else {
        showToast(`${t.failed}: ${device?.name || id}`, 'error', res.data.error);
      }
      load();
    } catch (e) {
      showToast(`${t.failed}: ${device?.name || id}`, 'error', e.response?.data?.detail || e.message);
    } finally {
      setBackingUp(null);
    }
  };

  const handleTest = async (id) => {
    const device = devices.find((d) => d.id === id);
    setTesting(id);
    try {
      const res = await api.post(`/devices/${id}/test`);
      if (res.data.status === 'success') {
        showToast(t.success, 'success');
      } else {
        showToast(`${t.failed}: ${device?.name || id}`, 'error', res.data.message);
      }
    } catch (e) {
      showToast(`${t.failed}: ${device?.name || id}`, 'error', e.response?.data?.detail || e.message);
    } finally {
      setTesting(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-800">{t.dev_title}</h1>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowGuide(true)} className="flex items-center gap-2 border border-gray-300 text-gray-600 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-50">
            <BookOpen size={16} /> {t.guide_title}
          </button>
          <button onClick={() => setShowBulkImport(true)} className="flex items-center gap-2 border border-gray-300 text-gray-600 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-50">
            <Upload size={16} /> {t.dev_csv_import}
          </button>
          <button onClick={() => setShowModal({})} className="flex items-center gap-2 bg-cyan-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-cyan-700">
            <Plus size={18} /> {t.dev_add}
          </button>
        </div>
      </div>

      {devices.length > 0 && (
        <div className="flex items-center gap-2">
          {[
            { id: 'all', label: t.dev_filter_all || 'All', color: 'bg-gray-100 text-gray-700 border-gray-300' },
            { id: 'fortigate', label: 'FortiGate', color: 'bg-orange-50 text-orange-700 border-orange-300' },
            { id: 'juniper', label: 'Juniper', color: 'bg-teal-50 text-teal-700 border-teal-300' },
            { id: 'cisco', label: 'Cisco', color: 'bg-indigo-50 text-indigo-700 border-indigo-300' },
            { id: 'paloalto', label: 'Palo Alto', color: 'bg-red-50 text-red-700 border-red-300' },
          ].map((v) => {
            const count = v.id === 'all' ? devices.length : devices.filter((d) => d.vendor === v.id).length;
            if (v.id !== 'all' && count === 0) return null;
            return (
              <button key={v.id} onClick={() => setVendorFilter(v.id)}
                className={`px-3 py-1.5 rounded-lg text-sm font-medium border transition-all ${vendorFilter === v.id ? v.color : 'bg-white text-gray-500 border-gray-200 hover:bg-gray-50'}`}>
                {v.label} <span className="ml-1 text-xs opacity-70">({count})</span>
              </button>
            );
          })}
        </div>
      )}

      {toast && (
        <div className={`p-3 rounded-lg text-sm ${toast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700 cursor-pointer'}`}
          onClick={() => { if (toast.detail) setErrorDetail(toast); }}>
          <div className="flex items-center justify-between">
            <span>{toast.msg}</span>
            {toast.detail && <span className="text-xs underline opacity-60 ml-2">{t.detail}</span>}
          </div>
        </div>
      )}

      {errorDetail && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-xl w-full max-w-lg">
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
              <h2 className="text-lg font-semibold text-red-600">{t.dev_error_title}</h2>
              <button onClick={() => setErrorDetail(null)} className="text-gray-400 hover:text-gray-600"><X size={20} /></button>
            </div>
            <div className="px-6 py-4">
              <p className="text-sm text-gray-600 mb-3">{errorDetail.msg}</p>
              <pre className="bg-gray-900 text-red-400 rounded-lg p-4 text-xs font-mono overflow-x-auto whitespace-pre-wrap">{errorDetail.detail}</pre>
            </div>
            <div className="px-6 py-3 border-t border-gray-100 flex justify-end">
              <button onClick={() => setErrorDetail(null)} className="px-4 py-2 bg-gray-100 text-gray-700 text-sm rounded-lg hover:bg-gray-200">{t.close}</button>
            </div>
          </div>
        </div>
      )}

      <div className="bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 text-gray-500 text-left">
            <tr>
              <th className="px-4 py-3 font-medium">{t.dev_name}</th>
              <th className="px-4 py-3 font-medium">{t.dev_vendor}</th>
              <th className="px-4 py-3 font-medium">{t.dev_ip}</th>
              <th className="px-4 py-3 font-medium">{t.dev_location}</th>
              <th className="px-4 py-3 font-medium">{t.dev_status}</th>
              <th className="px-4 py-3 font-medium">{t.dev_schedule}</th>
              <th className="px-4 py-3 font-medium">{t.dev_last_backup}</th>
              <th className="px-4 py-3 font-medium text-right">{t.actions}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {devices.filter((d) => vendorFilter === 'all' || d.vendor === vendorFilter).map((d) => (
              <tr key={d.id} className="hover:bg-gray-50">
                <td className="px-4 py-3 font-medium text-gray-800">{d.name}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${d.vendor === 'fortigate' ? 'bg-orange-100 text-orange-700' : d.vendor === 'cisco' ? 'bg-indigo-100 text-indigo-700' : d.vendor === 'paloalto' ? 'bg-red-100 text-red-700' : 'bg-teal-100 text-teal-700'}`}>
                    {d.vendor === 'fortigate' ? 'FortiGate' : d.vendor === 'cisco' ? (ciscoPlatforms[d.platform] || 'Cisco') : d.vendor === 'paloalto' ? 'Palo Alto' : 'Juniper'}
                  </span>
                  {d.vdom && <div className="text-xs text-gray-400 mt-0.5">VDOM: {d.vdom}</div>}
                </td>
                <td className="px-4 py-3 text-gray-600">{d.ip_address}:{d.port}</td>
                <td className="px-4 py-3 text-gray-600">
                  {d.location_name ? <span className="flex items-center gap-1 text-xs"><MapPin size={12} />{d.location_name}</span> : <span className="text-gray-400">-</span>}
                </td>
                <td className="px-4 py-3 text-gray-600">
                  <span className="text-green-600 font-medium">{d.backup_count}</span>
                  {d.failed_count > 0 && <span className="text-red-500 text-xs ml-1">/ {d.failed_count} {t.failed}</span>}
                </td>
                <td className="px-4 py-3 text-gray-600">{d.schedule_cron ? <span className="flex items-center gap-1"><Clock size={14} />{formatCron(d.schedule_cron, t)}</span> : <span className="text-gray-400">-</span>}</td>
                <td className="px-4 py-3 text-gray-600 text-xs">{d.last_backup ? new Date(d.last_backup).toLocaleString(locale) : t.dev_never}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center justify-end gap-1">
                    <button onClick={() => handleTest(d.id)} disabled={testing === d.id} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-cyan-600" title={t.dev_test_connection}>
                      <Wifi size={16} className={testing === d.id ? 'animate-pulse' : ''} />
                    </button>
                    <button onClick={() => handleBackup(d.id)} disabled={backingUp === d.id} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-green-600" title={t.dev_backup_now}>
                      <Play size={16} className={backingUp === d.id ? 'animate-spin' : ''} />
                    </button>
                    <button onClick={() => setShowModal(d)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-cyan-600" title={t.dev_edit}>
                      <Pencil size={16} />
                    </button>
                    <button onClick={() => setDeleteTarget(d)} className="p-1.5 rounded hover:bg-gray-100 text-gray-500 hover:text-red-600" title={t.dev_delete_remove}>
                      <Trash2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {devices.length === 0 && (
              <tr><td colSpan={8} className="px-4 py-8 text-center text-gray-400">{t.dev_no_devices}</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {showModal && <DeviceModal device={showModal.id ? showModal : null} onClose={() => setShowModal(null)} onSaved={() => { setShowModal(null); load(); }} t={t} />}
      {deleteTarget && <DeleteModal device={deleteTarget} onClose={() => setDeleteTarget(null)} onDeleted={() => { setDeleteTarget(null); load(); }} t={t} />}
      {showBulkImport && <BulkImportModal onClose={() => setShowBulkImport(false)} onImported={load} t={t} />}
      {showGuide && <VendorGuide onClose={() => setShowGuide(false)} t={t} />}
    </div>
  );
}
