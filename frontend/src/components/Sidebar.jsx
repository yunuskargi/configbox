import { NavLink } from 'react-router-dom';
import { LayoutDashboard, Server, Database, Settings, ChevronLeft, ChevronRight, LogOut, Box, MapPin, Users, Mail, Shield, Moon, Sun, Globe, CloudUpload, ExternalLink } from 'lucide-react';
import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useBranding } from '../context/BrandingContext';
import { useTheme } from '../context/ThemeContext';
import { useLang } from '../context/LangContext';

export default function Sidebar() {
  const [collapsed, setCollapsed] = useState(false);
  const { logout, user } = useAuth();
  const { appTitle } = useBranding();
  const { dark, toggle: toggleTheme } = useTheme();
  const { lang, t, switchLang } = useLang();

  const links = [
    { to: '/', label: t.sidebar_dashboard, icon: LayoutDashboard },
    { to: '/devices', label: t.sidebar_devices, icon: Server },
    { to: '/backups', label: t.sidebar_backups, icon: Database },
    { to: '/locations', label: t.sidebar_locations, icon: MapPin },
    ...(user?.role === 'admin' ? [
      { to: '/users', label: t.sidebar_users, icon: Users },
      { to: '/smtp', label: t.sidebar_mail, icon: Mail },
      { to: '/remote-backup', label: t.sidebar_remote_backup, icon: CloudUpload },
      { to: '/audit', label: t.sidebar_audit, icon: Shield },
    ] : []),
    { to: '/settings', label: t.sidebar_settings, icon: Settings },
  ];

  return (
    <aside className={`${collapsed ? 'w-[68px]' : 'w-60'} bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800 flex flex-col transition-all duration-200 min-h-screen`}>
      {/* Logo */}
      <div className={`flex items-center gap-2.5 h-16 border-b border-gray-100 dark:border-gray-800 ${collapsed ? 'px-3 justify-center' : 'px-4'}`}>
        <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-cyan-400 to-cyan-600 flex items-center justify-center shadow-md shadow-cyan-500/25 shrink-0">
          <Box className="w-5 h-5 text-white" />
        </div>
        {!collapsed && (
          <div className="min-w-0">
            <span className="text-base font-bold text-gray-800 dark:text-gray-100 block leading-tight tracking-tight">ConfigBox</span>
            {appTitle && <span className="text-[10px] text-gray-400 block truncate">{appTitle}</span>}
          </div>
        )}
        {!collapsed && (
          <button onClick={() => setCollapsed(true)} className="ml-auto p-1 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-800 transition">
            <ChevronLeft size={16} />
          </button>
        )}
      </div>

      {collapsed && (
        <button onClick={() => setCollapsed(false)} className="mx-auto mt-2 p-1.5 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:text-gray-300 dark:hover:bg-gray-800 transition">
          <ChevronRight size={16} />
        </button>
      )}

      {/* Nav */}
      <nav className="flex-1 py-4 space-y-0.5 px-3">
        {links.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            title={collapsed ? label : undefined}
            className={({ isActive }) =>
              `relative flex items-center gap-3 rounded-xl text-sm font-medium transition-all duration-150 ${collapsed ? 'justify-center px-0 py-2.5' : 'px-3 py-2.5'} ${
                isActive
                  ? 'bg-cyan-50 dark:bg-cyan-900/30 text-cyan-700 dark:text-cyan-400'
                  : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200'
              }`
            }
          >
            <Icon size={18} className="shrink-0" />
            {!collapsed && label}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="border-t border-gray-100 dark:border-gray-800 p-3">
        {!collapsed && (
          <div className="flex items-center gap-2.5 px-2 py-2 mb-1.5 rounded-xl bg-gray-50 dark:bg-gray-800/60">
            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-cyan-400 to-cyan-600 flex items-center justify-center text-white text-xs font-bold shrink-0 uppercase">
              {user?.username?.slice(0, 2)}
            </div>
            <div className="min-w-0">
              <p className="text-sm font-medium text-gray-700 dark:text-gray-200 truncate leading-tight">{user?.username}</p>
              <p className="text-[10px] text-gray-400">{user?.role === 'admin' ? 'Admin' : 'Backup Admin'}</p>
            </div>
          </div>
        )}
        <button onClick={() => switchLang(lang === 'tr' ? 'en' : 'tr')} title={collapsed ? (lang === 'tr' ? 'English' : 'Türkçe') : undefined} className={`flex items-center gap-3 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 w-full transition ${collapsed ? 'justify-center px-0' : 'px-3'}`}>
          <Globe size={17} className="shrink-0" />
          {!collapsed && <span>{lang === 'tr' ? 'English' : 'Türkçe'}</span>}
        </button>
        <button onClick={toggleTheme} title={collapsed ? (dark ? t.sidebar_light_mode : t.sidebar_dark_mode) : undefined} className={`flex items-center gap-3 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 w-full transition ${collapsed ? 'justify-center px-0' : 'px-3'}`}>
          {dark ? <Sun size={17} className="shrink-0 text-yellow-500" /> : <Moon size={17} className="shrink-0" />}
          {!collapsed && <span>{dark ? t.sidebar_light_mode : t.sidebar_dark_mode}</span>}
        </button>
        <button onClick={logout} title={collapsed ? t.sidebar_logout : undefined} className={`flex items-center gap-3 py-2 rounded-lg text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-600 w-full transition ${collapsed ? 'justify-center px-0' : 'px-3'}`}>
          <LogOut size={17} className="shrink-0" />
          {!collapsed && <span>{t.sidebar_logout}</span>}
        </button>
        {!collapsed && (
          <a
            href="https://github.com/yunuskargi/configbox"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-center gap-1 text-[11px] font-medium text-cyan-600 dark:text-cyan-400 hover:text-cyan-700 dark:hover:text-cyan-300 hover:underline mt-2 transition"
          >
            <span>GitHub</span>
            <ExternalLink size={10} />
          </a>
        )}
      </div>
    </aside>
  );
}
