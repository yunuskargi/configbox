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
    <aside className={`${collapsed ? 'w-16' : 'w-60'} bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700 flex flex-col transition-all duration-200 min-h-screen`}>
      <div className="flex items-center gap-2 px-4 h-16 border-b border-gray-200 dark:border-gray-700">
        <Box className="w-7 h-7 text-cyan-600 shrink-0" />
        {!collapsed && (
          <div className="min-w-0">
            <span className="text-lg font-bold text-gray-800 dark:text-gray-100 block leading-tight">ConfigBox</span>
            {appTitle && <span className="text-[10px] text-gray-400 block truncate">{appTitle}</span>}
          </div>
        )}
        <button onClick={() => setCollapsed(!collapsed)} className="ml-auto text-gray-400 hover:text-gray-600 dark:hover:text-gray-300">
          {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
        </button>
      </div>

      <nav className="flex-1 py-4 space-y-1 px-2">
        {links.map(({ to, label, icon: Icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                isActive ? 'bg-cyan-50 dark:bg-cyan-900/30 text-cyan-700 dark:text-cyan-400' : 'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-gray-200'
              }`
            }
          >
            <Icon size={20} className="shrink-0" />
            {!collapsed && label}
          </NavLink>
        ))}
      </nav>

      <div className="border-t border-gray-200 dark:border-gray-700 p-3">
        {!collapsed && (
          <div className="px-3 py-1.5 mb-1">
            <p className="text-sm font-medium text-gray-700 dark:text-gray-200">{user?.username}</p>
            <p className="text-xs text-gray-400">{user?.role === 'admin' ? 'Admin' : 'Backup Admin'}</p>
          </div>
        )}
        <button onClick={() => switchLang(lang === 'tr' ? 'en' : 'tr')} className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 w-full">
          <Globe size={20} className="shrink-0" />
          {!collapsed && <span>{lang === 'tr' ? 'English' : 'Türkçe'}</span>}
        </button>
        <button onClick={toggleTheme} className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 w-full">
          {dark ? <Sun size={20} className="shrink-0 text-yellow-500" /> : <Moon size={20} className="shrink-0" />}
          {!collapsed && <span>{dark ? t.sidebar_light_mode : t.sidebar_dark_mode}</span>}
        </button>
        <button onClick={logout} className="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 hover:text-red-700 w-full">
          <LogOut size={20} className="shrink-0" />
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
