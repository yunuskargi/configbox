import { useEffect, useState } from 'react';
import api from '../api/client';
import { Cloud, HardDrive, Send, Info, CheckCircle, ExternalLink, BookOpen } from 'lucide-react';
import { useLang } from '../context/LangContext';

export default function RemoteBackup() {
  const [s3, setS3] = useState({ s3_enabled: false, s3_endpoint: '', s3_region: 'us-east-1', s3_bucket: '', s3_access_key: '', s3_secret_key: '', s3_use_ssl: true, s3_prefix: '' });
  const [gdrive, setGDrive] = useState({ gdrive_enabled: false, gdrive_client_id: '', gdrive_client_secret: '', gdrive_folder_id: '', gdrive_authorized: false });
  const [authCode, setAuthCode] = useState('');
  const [saving, setSaving] = useState(false);
  const [testingS3, setTestingS3] = useState(false);
  const [testingGDrive, setTestingGDrive] = useState(false);
  const [authorizing, setAuthorizing] = useState(false);
  const [toast, setToast] = useState(null);
  const { t } = useLang();

  useEffect(() => {
    api.get('/settings/s3').then((r) => setS3(r.data));
    api.get('/settings/gdrive').then((r) => setGDrive(r.data));
  }, []);

  const showToast = (msg, type = 'success') => {
    setToast({ msg, type });
    setTimeout(() => setToast(null), 4000);
  };

  const saveS3 = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put('/settings/s3', s3);
      showToast(t.rb_saved);
    } catch {
      showToast(t.error_occurred, 'error');
    } finally {
      setSaving(false);
    }
  };

  const saveGDrive = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      await api.put('/settings/gdrive', gdrive);
      showToast(t.rb_saved);
    } catch {
      showToast(t.error_occurred, 'error');
    } finally {
      setSaving(false);
    }
  };

  const testS3 = async () => {
    setTestingS3(true);
    try {
      await api.post('/settings/s3/test');
      showToast(t.rb_test_ok);
    } catch (err) {
      showToast(err.response?.data?.detail || t.rb_test_fail, 'error');
    } finally {
      setTestingS3(false);
    }
  };

  const authorizeGDrive = async () => {
    setAuthorizing(true);
    try {
      const res = await api.post('/settings/gdrive/auth-url');
      window.open(res.data.url, '_blank');
    } catch (err) {
      showToast(err.response?.data?.detail || t.error_occurred, 'error');
    } finally {
      setAuthorizing(false);
    }
  };

  const submitAuthCode = async () => {
    if (!authCode.trim()) return;
    setAuthorizing(true);
    try {
      await api.post('/settings/gdrive/callback', { code: authCode.trim() });
      setGDrive({ ...gdrive, gdrive_authorized: true });
      setAuthCode('');
      showToast(t.rb_gdrive_auth_ok);
    } catch (err) {
      showToast(err.response?.data?.detail || t.rb_gdrive_auth_fail, 'error');
    } finally {
      setAuthorizing(false);
    }
  };

  const testGDrive = async () => {
    setTestingGDrive(true);
    try {
      await api.post('/settings/gdrive/test');
      showToast(t.rb_test_ok);
    } catch (err) {
      showToast(err.response?.data?.detail || t.rb_test_fail, 'error');
    } finally {
      setTestingGDrive(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-800 dark:text-gray-100">{t.rb_title}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t.rb_description}</p>
      </div>

      {toast && (
        <div className={`p-3 rounded-lg text-sm ${toast.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
          {toast.msg}
        </div>
      )}

      {/* S3 */}
      <div className="flex gap-4">
      <div className="flex-1 max-w-2xl bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 flex items-center gap-2"><HardDrive size={20} />{t.rb_s3_title}</h2>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={s3.s3_enabled} onChange={(e) => setS3({ ...s3, s3_enabled: e.target.checked })} className="rounded" />
            <span className={s3.s3_enabled ? 'text-green-600 font-medium' : 'text-gray-400'}>{s3.s3_enabled ? t.rb_enabled : t.rb_disabled}</span>
          </label>
        </div>

        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-3 mb-4">
          <p className="text-xs text-blue-700 dark:text-blue-300 flex items-start gap-2"><Info size={14} className="shrink-0 mt-0.5" />{t.rb_s3_info}</p>
        </div>

        <form onSubmit={saveS3} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_endpoint}</label>
              <input value={s3.s3_endpoint} onChange={(e) => setS3({ ...s3, s3_endpoint: e.target.value })} placeholder="s3.amazonaws.com" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_region} <span className="text-gray-400 font-normal">({t.optional})</span></label>
              <input value={s3.s3_region} onChange={(e) => setS3({ ...s3, s3_region: e.target.value })} placeholder="us-east-1" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_bucket}</label>
              <input value={s3.s3_bucket} onChange={(e) => setS3({ ...s3, s3_bucket: e.target.value })} placeholder="my-backups" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_prefix} <span className="text-gray-400 font-normal">({t.optional})</span></label>
              <input value={s3.s3_prefix} onChange={(e) => setS3({ ...s3, s3_prefix: e.target.value })} placeholder="configbox/" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_access_key}</label>
              <input type="password" value={s3.s3_access_key} onChange={(e) => setS3({ ...s3, s3_access_key: e.target.value })} className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_s3_secret_key}</label>
              <input type="password" value={s3.s3_secret_key} onChange={(e) => setS3({ ...s3, s3_secret_key: e.target.value })} className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
          </div>
          <label className="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input type="checkbox" checked={s3.s3_use_ssl} onChange={(e) => setS3({ ...s3, s3_use_ssl: e.target.checked })} className="rounded" />
            {t.rb_s3_use_ssl}
          </label>
          <div className="flex gap-3">
            <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{t.save}</button>
            <button type="button" onClick={testS3} disabled={testingS3} className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 dark:text-gray-300">
              <Send size={14} />{testingS3 ? t.rb_testing : t.rb_test}
            </button>
          </div>
        </form>
      </div>
      <div className="w-72 shrink-0 hidden lg:block">
        <div className="bg-white dark:bg-gray-800 rounded-xl p-4 shadow-sm border border-gray-100 dark:border-gray-700 space-y-3">
          <h4 className="text-xs font-semibold text-cyan-700 dark:text-cyan-400 flex items-center gap-1.5"><BookOpen size={14} />{t.rb_guide_s3}</h4>
          <ol className="space-y-2 text-xs text-gray-600 dark:text-gray-400">
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">1</span>{t.rb_guide_s3_step1}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">2</span>{t.rb_guide_s3_step2}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">3</span>{t.rb_guide_s3_step3}</li>
          </ol>
        </div>
      </div>
      </div>

      {/* Google Drive */}
      <div className="flex gap-4">
      <div className="flex-1 max-w-2xl bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-100 dark:border-gray-700">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 flex items-center gap-2"><Cloud size={20} />{t.rb_gdrive_title}</h2>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={gdrive.gdrive_enabled} onChange={(e) => setGDrive({ ...gdrive, gdrive_enabled: e.target.checked })} className="rounded" />
            <span className={gdrive.gdrive_enabled ? 'text-green-600 font-medium' : 'text-gray-400'}>{gdrive.gdrive_enabled ? t.rb_enabled : t.rb_disabled}</span>
          </label>
        </div>

        <form onSubmit={saveGDrive} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_gdrive_client_id}</label>
              <input value={gdrive.gdrive_client_id} onChange={(e) => setGDrive({ ...gdrive, gdrive_client_id: e.target.value })} placeholder="xxxx.apps.googleusercontent.com" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100 text-xs" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_gdrive_client_secret}</label>
              <input type="password" value={gdrive.gdrive_client_secret} onChange={(e) => setGDrive({ ...gdrive, gdrive_client_secret: e.target.value })} className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            </div>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{t.rb_gdrive_folder_id}</label>
            <input value={gdrive.gdrive_folder_id} onChange={(e) => setGDrive({ ...gdrive, gdrive_folder_id: e.target.value })} placeholder="1AbC2dEfGhIjKlMnOpQrStUvWxYz" className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100" />
            <p className="text-xs text-gray-400 mt-1">{t.rb_gdrive_folder_hint}</p>
          </div>

          <div className="flex gap-3">
            <button type="submit" disabled={saving} className="px-4 py-2 bg-cyan-600 text-white text-sm rounded-lg hover:bg-cyan-700 disabled:opacity-50">{t.save}</button>
          </div>
        </form>

        <hr className="my-4 border-gray-200 dark:border-gray-700" />

        {gdrive.gdrive_authorized ? (
          <div className="flex items-center gap-2 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg mb-4">
            <CheckCircle size={16} className="text-green-600" />
            <span className="text-sm text-green-700 dark:text-green-300 font-medium">{t.rb_gdrive_authorized}</span>
          </div>
        ) : (
          <div className="space-y-3">
            <p className="text-sm text-gray-600 dark:text-gray-400">{t.rb_gdrive_auth_steps}</p>
            <div className="flex gap-3">
              <button onClick={authorizeGDrive} disabled={authorizing || !gdrive.gdrive_client_id} className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 disabled:opacity-50">
                <ExternalLink size={14} />{t.rb_gdrive_authorize}
              </button>
            </div>
            <div className="flex gap-2">
              <input value={authCode} onChange={(e) => setAuthCode(e.target.value)} placeholder={t.rb_gdrive_paste_code} className="flex-1 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:ring-2 focus:ring-cyan-500 dark:bg-gray-700 dark:text-gray-100 text-sm" />
              <button onClick={submitAuthCode} disabled={authorizing || !authCode.trim()} className="px-4 py-2 bg-green-600 text-white text-sm rounded-lg hover:bg-green-700 disabled:opacity-50">
                {authorizing ? t.rb_testing : t.rb_gdrive_confirm}
              </button>
            </div>
          </div>
        )}

        {gdrive.gdrive_authorized && (
          <div className="flex gap-3">
            <button onClick={testGDrive} disabled={testingGDrive} className="flex items-center gap-2 px-4 py-2 border border-gray-300 dark:border-gray-600 text-sm rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 dark:text-gray-300">
              <Send size={14} />{testingGDrive ? t.rb_testing : t.rb_test}
            </button>
            <button onClick={() => { setGDrive({ ...gdrive, gdrive_authorized: false }); authorizeGDrive(); }} className="flex items-center gap-2 px-4 py-2 text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400">
              <ExternalLink size={14} />{t.rb_gdrive_reauthorize}
            </button>
          </div>
        )}
      </div>
      <div className="w-72 shrink-0 hidden lg:block">
        <div className="bg-white dark:bg-gray-800 rounded-xl p-4 shadow-sm border border-gray-100 dark:border-gray-700 space-y-3">
          <h4 className="text-xs font-semibold text-cyan-700 dark:text-cyan-400 flex items-center gap-1.5"><BookOpen size={14} />{t.rb_guide_gdrive}</h4>
          <ol className="space-y-2 text-xs text-gray-600 dark:text-gray-400">
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">1</span>{t.rb_guide_gdrive_step1}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">2</span>{t.rb_guide_gdrive_step2}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">3</span>{t.rb_guide_gdrive_step3}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">4</span>{t.rb_guide_gdrive_step4}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">5</span>{t.rb_guide_gdrive_step5}</li>
            <li className="flex gap-2"><span className="flex items-center justify-center w-4 h-4 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px] font-bold shrink-0 mt-0.5">6</span>{t.rb_guide_gdrive_step6}</li>
          </ol>
          <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-2.5">
            <p className="text-[11px] text-amber-800 dark:text-amber-300"><strong>{t.rb_guide_note_title}:</strong> {t.rb_guide_note}</p>
          </div>
        </div>
      </div>
      </div>
    </div>
  );
}
