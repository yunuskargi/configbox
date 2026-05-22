import { useEffect, useState } from 'react';
import api from '../api/client';
import { Mail, Bell, Send } from 'lucide-react';
import { useLang } from '../context/LangContext';

export default function SmtpSettings() {
  const [smtp, setSmtp] = useState({ smtp_host: '', smtp_port: 587, smtp_username: '', smtp_password: '', smtp_use_tls: true, smtp_from_email: '', smtp_from_name: 'ConfigBox' });
  const [notify, setNotify] = useState({ notify_on_success: false, notify_on_failure: true, notify_on_change: false, notify_daily_summary: false, notify_recipients: '' });
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [toast, setToast] = useState(null);
  const { t } = useLang();

  useEffect(() => {
    api.get('/settings/smtp').then((r) => setSmtp(r.data));
    api.get('/settings/notify').then((r) => setNotify(r.data));
  }, []);

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const saveSmtp = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put('/settings/smtp', smtp);
      showToast(t.smtp_saved);
    } catch {
      showToast(t.error_occurred, 'error');
    } finally {
      setSaving(false);
    }
  };

  const saveNotify = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put('/settings/notify', notify);
      showToast(t.smtp_notify_saved);
    } catch {
      showToast(t.error_occurred, 'error');
    } finally {
      setSaving(false);
    }
  };

  const testSmtp = async () => {
    setTesting(true);
    try {
      await api.post('/settings/smtp/test');
      showToast(t.smtp_test_ok);
    } catch (err) {
      showToast(err.response?.data?.detail || t.smtp_test_fail, 'error');
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold text-gray-800">{t.smtp_title}</h1>

      {toast && (
        <div className={`p-3 rounded-lg text-sm ${toast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
          {toast.msg}
        </div>
      )}

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2"><Mail size={20} />{t.smtp_server}</h2>
        <form onSubmit={saveSmtp} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_host}</label>
              <input value={smtp.smtp_host} onChange={(e) => setSmtp({ ...smtp, smtp_host: e.target.value })} placeholder="smtp.gmail.com" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_port}</label>
              <input type="number" value={smtp.smtp_port} onChange={(e) => setSmtp({ ...smtp, smtp_port: Number(e.target.value) })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_username}</label>
              <input value={smtp.smtp_username} onChange={(e) => setSmtp({ ...smtp, smtp_username: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_password}</label>
              <input type="password" value={smtp.smtp_password} onChange={(e) => setSmtp({ ...smtp, smtp_password: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_from_email}</label>
              <input value={smtp.smtp_from_email} onChange={(e) => setSmtp({ ...smtp, smtp_from_email: e.target.value })} placeholder="configbox@company.com" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_from_name}</label>
              <input value={smtp.smtp_from_name} onChange={(e) => setSmtp({ ...smtp, smtp_from_name: e.target.value })} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            </div>
          </div>
          <label className="flex items-center gap-2 text-sm text-gray-700">
            <input type="checkbox" checked={smtp.smtp_use_tls} onChange={(e) => setSmtp({ ...smtp, smtp_use_tls: e.target.checked })} className="rounded" />
            {t.smtp_use_tls}
          </label>
          <div className="flex gap-3">
            <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{t.save}</button>
            <button type="button" onClick={testSmtp} disabled={testing} className="flex items-center gap-2 px-4 py-2 border border-gray-300 text-sm rounded-lg hover:bg-gray-50 disabled:opacity-50">
              <Send size={14} />{testing ? t.smtp_testing : t.smtp_test}
            </button>
          </div>
        </form>
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2"><Bell size={20} />{t.smtp_notify_title}</h2>
        <form onSubmit={saveNotify} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.smtp_recipients}</label>
            <input value={notify.notify_recipients} onChange={(e) => setNotify({ ...notify, notify_recipients: e.target.value })} placeholder="admin@company.com, it@company.com" className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            <p className="text-xs text-gray-400 mt-1">{t.smtp_recipients_desc}</p>
          </div>
          <div className="space-y-2">
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={notify.notify_on_success} onChange={(e) => setNotify({ ...notify, notify_on_success: e.target.checked })} className="rounded" />
              {t.smtp_notify_success}
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={notify.notify_on_failure} onChange={(e) => setNotify({ ...notify, notify_on_failure: e.target.checked })} className="rounded" />
              {t.smtp_notify_failure}
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={notify.notify_on_change} onChange={(e) => setNotify({ ...notify, notify_on_change: e.target.checked })} className="rounded" />
              {t.smtp_notify_change}
            </label>
            <label className="flex items-center gap-2 text-sm text-gray-700">
              <input type="checkbox" checked={notify.notify_daily_summary} onChange={(e) => setNotify({ ...notify, notify_daily_summary: e.target.checked })} className="rounded" />
              {t.smtp_notify_summary}
            </label>
          </div>
          <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{t.save}</button>
        </form>
      </div>
    </div>
  );
}
