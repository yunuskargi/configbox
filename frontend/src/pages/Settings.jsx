import { useEffect, useState } from 'react';
import api from '../api/client';
import { useAuth } from '../context/AuthContext';
import { useBranding } from '../context/BrandingContext';
import { useLang } from '../context/LangContext';
import { Save, Key, HardDrive, FolderOpen, Clock, Info, Palette, ShieldCheck, ShieldOff, Loader2, CheckCircle, AlertTriangle } from 'lucide-react';

export default function Settings() {
  const { user, refreshUser } = useAuth();
  const [settings, setSettings] = useState({ backup_dir: '', retention_days: 90, app_title: '', archive_enabled: false, archive_after_days: 30 });
  const { refresh: refreshBranding } = useBranding();
  const [passwords, setPasswords] = useState({ current_password: '', new_password: '', confirm: '' });
  const [saving, setSaving] = useState(false);
  const [toast, setToast] = useState(null);
  const { t } = useLang();

  const [twoFA, setTwoFA] = useState({ step: 'idle', qrCode: null, secret: null, code: '', disablePass: '', loading: false });

  useEffect(() => {
    api.get('/settings').then((r) => setSettings(r.data));
  }, []);

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 3000);
  };

  const saveSettings = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put('/settings', settings);
      refreshBranding();
      showToast(t.set_saved);
    } catch {
      showToast(t.error_occurred, 'error');
    } finally {
      setSaving(false);
    }
  };

  const changePassword = async (e) => {
    e.preventDefault();
    if (passwords.new_password.length < 8 || !/(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/.test(passwords.new_password)) {
      showToast(t.set_password_policy, 'error');
      return;
    }
    if (passwords.new_password !== passwords.confirm) {
      showToast(t.set_password_mismatch, 'error');
      return;
    }
    try {
      await api.post('/auth/change-password', {
        current_password: passwords.current_password,
        new_password: passwords.new_password,
      });
      showToast(t.set_password_changed);
      setPasswords({ current_password: '', new_password: '', confirm: '' });
      // Refresh user context to clear must_change_password flag
      await refreshUser();
    } catch (err) {
      showToast(err.response?.data?.detail || t.error_occurred, 'error');
    }
  };

  return (
    <div className="space-y-6 max-w-2xl">
      <h1 className="text-2xl font-bold text-gray-800">{t.set_title}</h1>

      {user?.must_change_password && (
        <div className="flex items-center gap-3 px-4 py-3 bg-amber-50 border border-amber-200 rounded-xl text-amber-700 text-sm font-medium">
          <AlertTriangle size={18} className="text-amber-500 shrink-0" />
          {t.set_must_change_password || 'You must change your password before continuing.'}
        </div>
      )}

      {toast && (
        <div className="fixed top-6 right-6 z-50 animate-[slideIn_0.3s_ease-out]">
          <div className={`flex items-center gap-3 px-4 py-3 rounded-xl shadow-lg border text-sm font-medium ${toast.type === 'success' ? 'bg-white border-green-200 text-green-700' : 'bg-white border-red-200 text-red-700'}`}>
            {toast.type === 'success' ? <CheckCircle size={18} className="text-green-500 shrink-0" /> : <Info size={18} className="text-red-500 shrink-0" />}
            {toast.msg}
          </div>
        </div>
      )}

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2"><Palette size={20} />{t.set_personalize}</h2>
        <form onSubmit={saveSettings} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_company_title}</label>
            <input value={settings.app_title} onChange={(e) => setSettings((s) => ({ ...s, app_title: e.target.value }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            <p className="text-xs text-gray-400 mt-1">{t.set_company_desc}</p>
          </div>
          <button type="submit" disabled={saving} className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">
            <Save size={14} /> {saving ? t.saving : t.save}
          </button>
        </form>
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2"><HardDrive size={20} />{t.set_backup_settings}</h2>
        <form onSubmit={saveSettings} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1 flex items-center gap-1">
              <FolderOpen size={14} /> {t.set_backup_dir}
            </label>
            <input value={settings.backup_dir} onChange={(e) => setSettings((s) => ({ ...s, backup_dir: e.target.value }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 font-mono text-sm" />
            <div className="mt-2 bg-cyan-50 border border-cyan-100 rounded-lg p-3">
              <div className="flex gap-2">
                <Info size={14} className="text-cyan-500 shrink-0 mt-0.5" />
                <div className="text-xs text-cyan-700 space-y-1">
                  <p>{t.set_backup_dir_info}</p>
                  <code className="block bg-cyan-100 rounded px-2 py-1 text-[11px]">
                    {t.set_backup_dir_structure(settings.backup_dir)}
                  </code>
                  <p>{t.set_backup_dir_docker}</p>
                  <p className="text-cyan-500">{t.set_backup_dir_default}</p>
                </div>
              </div>
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1 flex items-center gap-1">
              <Clock size={14} /> {t.set_retention}
            </label>
            <input type="number" min="1" max="3650" value={settings.retention_days} onChange={(e) => setSettings((s) => ({ ...s, retention_days: Number(e.target.value) }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
            <p className="text-xs text-gray-400 mt-1">{t.set_retention_desc}</p>
          </div>
          <div>
            <label className="flex items-center gap-2 text-sm font-medium text-gray-700 mb-2">
              <input type="checkbox" checked={settings.archive_enabled} onChange={(e) => setSettings((s) => ({ ...s, archive_enabled: e.target.checked }))} className="rounded" />
              {t.set_archive}
            </label>
            {settings.archive_enabled && (
              <div className="ml-6">
                <input type="number" min="1" max="3650" value={settings.archive_after_days} onChange={(e) => setSettings((s) => ({ ...s, archive_after_days: Number(e.target.value) }))} className="w-24 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" />
                <span className="text-sm text-gray-500 ml-2">{t.set_archive_days}</span>
                <p className="text-xs text-gray-400 mt-1">{t.set_archive_desc}</p>
              </div>
            )}
          </div>
          <button type="submit" disabled={saving} className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">
            <Save size={14} /> {saving ? t.saving : t.save}
          </button>
        </form>
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2"><Key size={20} />{t.set_change_password}</h2>
        <form onSubmit={changePassword} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_current_password}</label>
            <input type="password" value={passwords.current_password} onChange={(e) => setPasswords((p) => ({ ...p, current_password: e.target.value }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_new_password}</label>
            <input type="password" value={passwords.new_password} onChange={(e) => setPasswords((p) => ({ ...p, new_password: e.target.value }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_confirm_password}</label>
            <input type="password" value={passwords.confirm} onChange={(e) => setPasswords((p) => ({ ...p, confirm: e.target.value }))} className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500" required />
          </div>
          <button type="submit" className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700">
            <Key size={14} /> {t.set_change_password}
          </button>
        </form>
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <h2 className="text-lg font-semibold text-gray-800 mb-4 flex items-center gap-2">
          <ShieldCheck size={20} />{t.set_2fa_title}
        </h2>

        {user?.totp_enabled ? (
          <div className="space-y-4">
            <div className="flex items-center gap-2 text-green-600 bg-green-50 px-4 py-3 rounded-lg border border-green-100">
              <ShieldCheck size={18} />
              <span className="text-sm font-medium">{t.set_2fa_active}</span>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_2fa_disable_prompt}</label>
              <input
                type="password"
                value={twoFA.disablePass}
                onChange={(e) => setTwoFA((s) => ({ ...s, disablePass: e.target.value }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
                placeholder={t.set_current_password}
              />
            </div>
            <button
              onClick={async () => {
                if (!twoFA.disablePass) return;
                setTwoFA((s) => ({ ...s, loading: true }));
                try {
                  await api.post('/auth/2fa/disable', { current_password: twoFA.disablePass, new_password: twoFA.disablePass });
                  showToast(t.set_2fa_disabled);
                  setTwoFA({ step: 'idle', qrCode: null, secret: null, code: '', disablePass: '', loading: false });
                  window.location.reload();
                } catch (err) {
                  showToast(err.response?.data?.detail || t.error_occurred, 'error');
                  setTwoFA((s) => ({ ...s, loading: false }));
                }
              }}
              disabled={twoFA.loading || !twoFA.disablePass}
              className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white text-sm rounded-lg hover:bg-red-700 disabled:opacity-50"
            >
              <ShieldOff size={14} /> {twoFA.loading ? t.set_2fa_processing : t.set_2fa_disable}
            </button>
          </div>
        ) : twoFA.step === 'idle' ? (
          <div className="space-y-4">
            <p className="text-sm text-gray-500">{t.set_2fa_desc}</p>
            <button
              onClick={async () => {
                setTwoFA((s) => ({ ...s, loading: true }));
                try {
                  const res = await api.post('/auth/2fa/setup');
                  setTwoFA({ step: 'setup', qrCode: res.data.qr_code, secret: res.data.secret, code: '', disablePass: '', loading: false });
                } catch (err) {
                  showToast(err.response?.data?.detail || t.error_occurred, 'error');
                  setTwoFA((s) => ({ ...s, loading: false }));
                }
              }}
              disabled={twoFA.loading}
              className="flex items-center gap-2 px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50"
            >
              {twoFA.loading ? <Loader2 size={14} className="animate-spin" /> : <ShieldCheck size={14} />}
              {t.set_2fa_enable}
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm text-gray-500">{t.set_2fa_scan}</p>
            <div className="flex justify-center bg-gray-50 rounded-lg p-4 border border-gray-200">
              <img src={twoFA.qrCode} alt="2FA QR Code" className="w-48 h-48" />
            </div>
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <p className="text-xs text-gray-500 mb-1">{t.set_2fa_manual}</p>
              <code className="text-sm font-mono text-gray-700 select-all">{twoFA.secret}</code>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">{t.set_2fa_code}</label>
              <input
                type="text"
                value={twoFA.code}
                onChange={(e) => setTwoFA((s) => ({ ...s, code: e.target.value.replace(/\D/g, '').slice(0, 6) }))}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 text-center text-lg tracking-widest font-mono"
                placeholder="000000"
                maxLength={6}
              />
            </div>
            <div className="flex gap-2">
              <button
                onClick={async () => {
                  if (twoFA.code.length !== 6) return;
                  setTwoFA((s) => ({ ...s, loading: true }));
                  try {
                    await api.post('/auth/2fa/verify', { code: twoFA.code });
                    showToast(t.set_2fa_activated);
                    setTwoFA({ step: 'idle', qrCode: null, secret: null, code: '', disablePass: '', loading: false });
                    window.location.reload();
                  } catch (err) {
                    showToast(err.response?.data?.detail || t.set_2fa_invalid, 'error');
                    setTwoFA((s) => ({ ...s, loading: false }));
                  }
                }}
                disabled={twoFA.loading || twoFA.code.length !== 6}
                className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 disabled:opacity-50"
              >
                {twoFA.loading ? <Loader2 size={14} className="animate-spin" /> : <ShieldCheck size={14} />}
                {t.set_2fa_verify}
              </button>
              <button
                onClick={() => setTwoFA({ step: 'idle', qrCode: null, secret: null, code: '', disablePass: '', loading: false })}
                className="px-4 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50"
              >
                {t.cancel}
              </button>
            </div>
          </div>
        )}
      </div>

      <div className="bg-gray-50 rounded-xl p-4 border border-gray-200">
        <h3 className="text-sm font-medium text-gray-600 mb-2">{t.set_reset_title}</h3>
        <p className="text-xs text-gray-400">{t.set_reset_desc}</p>
        <code className="block bg-gray-100 rounded px-3 py-2 text-xs font-mono text-gray-600 mt-2">
          docker compose exec backend /confbox reset-password admin newpassword
        </code>
      </div>
    </div>
  );
}
