import { useState } from 'react';
import { useNavigate, Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useBranding } from '../context/BrandingContext';
import { useLang } from '../context/LangContext';
import { Box, Eye, EyeOff } from 'lucide-react';

export default function Login() {
  const { login, user } = useAuth();
  const { appTitle } = useBranding();
  const { t } = useLang();
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showPass, setShowPass] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [needs2FA, setNeeds2FA] = useState(false);
  const [totpCode, setTotpCode] = useState('');

  if (user) return <Navigate to="/" />;

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const result = await login(username, password, needs2FA ? totpCode : undefined);
      if (result?.requires_2fa) {
        setNeeds2FA(true);
        setLoading(false);
        return;
      }
      navigate('/');
    } catch (err) {
      setError(err.response?.data?.detail || t.error_occurred);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex bg-gradient-to-br from-slate-900 via-cyan-950 to-slate-900">
      <div className="hidden lg:flex lg:w-1/2 items-center justify-center p-12">
        <div className="max-w-md text-center">
          <div className="flex items-center justify-center gap-3 mb-8">
            <div className="w-16 h-16 bg-cyan-500/20 rounded-2xl flex items-center justify-center backdrop-blur-sm border border-cyan-400/20">
              <Box className="w-9 h-9 text-cyan-400" />
            </div>
          </div>
          <h1 className="text-4xl font-bold text-white mb-4">ConfBox</h1>
          {appTitle && <p className="text-cyan-300/80 text-base font-medium mb-2">{appTitle}</p>}
          <p className="text-cyan-200/70 text-lg leading-relaxed">{t.login_tagline}</p>
          <p className="text-cyan-300/50 text-sm mt-2">{t.login_description}</p>
          <div className="mt-12 grid grid-cols-3 gap-6 text-center">
            <div>
              <div className="text-2xl font-bold text-white">7/24</div>
              <div className="text-xs text-cyan-300/50 mt-1">{t.login_auto_backup}</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-white">.conf</div>
              <div className="text-xs text-cyan-300/50 mt-1">{t.login_plain_file}</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-white">REST</div>
              <div className="text-xs text-cyan-300/50 mt-1">{t.login_api_support}</div>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <div className="lg:hidden flex items-center justify-center gap-2 mb-8">
            <Box className="w-8 h-8 text-cyan-400" />
            <h1 className="text-2xl font-bold text-white">ConfBox</h1>
          </div>

          <div className="bg-white/5 backdrop-blur-xl rounded-2xl p-8 border border-white/10 shadow-2xl">
            {!needs2FA ? (
              <>
                <h2 className="text-xl font-semibold text-white mb-1">{t.login_welcome}</h2>
                <p className="text-sm text-cyan-200/50 mb-6">{t.login_subtitle}</p>
              </>
            ) : (
              <>
                <h2 className="text-xl font-semibold text-white mb-1">{t.login_2fa_title}</h2>
                <p className="text-sm text-cyan-200/50 mb-6">{t.login_2fa_subtitle}</p>
              </>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="text-sm text-red-300 bg-red-500/10 border border-red-500/20 p-3 rounded-xl">
                  {error}
                </div>
              )}
              {!needs2FA ? (
                <>
                  <div>
                    <label className="block text-sm font-medium text-cyan-200/70 mb-1.5">{t.login_username}</label>
                    <input
                      type="text"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      className="login-input w-full px-4 py-2.5 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/20 focus:outline-none focus:ring-2 focus:ring-cyan-500/50 focus:border-cyan-500/50 transition-all"
                      placeholder="admin"
                      required
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-cyan-200/70 mb-1.5">{t.login_password}</label>
                    <div className="relative">
                      <input
                        type={showPass ? 'text' : 'password'}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="login-input w-full px-4 py-2.5 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/20 focus:outline-none focus:ring-2 focus:ring-cyan-500/50 focus:border-cyan-500/50 transition-all pr-10"
                        placeholder="••••••••"
                        required
                      />
                      <button type="button" onClick={() => setShowPass(!showPass)} className="absolute right-3 top-1/2 -translate-y-1/2 text-white/30 hover:text-white/60">
                        {showPass ? <EyeOff size={18} /> : <Eye size={18} />}
                      </button>
                    </div>
                  </div>
                </>
              ) : (
                <div>
                  <label className="block text-sm font-medium text-cyan-200/70 mb-1.5">{t.login_2fa_code}</label>
                  <input
                    type="text"
                    value={totpCode}
                    onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                    className="login-input w-full px-4 py-2.5 bg-white/5 border border-white/10 rounded-xl text-white placeholder-white/20 focus:outline-none focus:ring-2 focus:ring-cyan-500/50 focus:border-cyan-500/50 transition-all text-center text-lg tracking-widest font-mono"
                    placeholder="000000"
                    maxLength={6}
                    autoFocus
                  />
                </div>
              )}
              <button
                type="submit"
                disabled={loading}
                className="w-full py-3 bg-gradient-to-r from-cyan-500 to-cyan-600 text-white font-semibold rounded-xl hover:from-cyan-600 hover:to-cyan-700 transition-all shadow-lg shadow-cyan-500/25 disabled:opacity-50"
              >
                {loading ? t.login_logging_in : needs2FA ? t.login_2fa_verify : t.login_button}
              </button>
            </form>
          </div>

          <p className="text-center text-cyan-300/30 text-xs mt-6">{t.login_footer}</p>
        </div>
      </div>
    </div>
  );
}
