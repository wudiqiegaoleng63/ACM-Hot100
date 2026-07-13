import { Flame, ListChecks, LogIn, LogOut, Send, User } from 'lucide-react';
import { Link, NavLink, Outlet, useMatch } from 'react-router';

import { useAuth } from '@/features/auth/contexts/auth-context';

export default function RootLayout() {
  const { user, isLoading, logout } = useAuth();
  const isProblemWorkspace = Boolean(useMatch('/problems/:slug'));

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--surface)]">
        <div className="mx-auto flex h-14 max-w-[1200px] items-center justify-between px-4">
          <Link className="flex items-center gap-2 text-base font-bold no-underline text-[var(--accent)] sm:text-lg" to="/">
            <Flame aria-hidden="true" size={22} />
            ACM HOT 100
          </Link>
          <nav className="flex items-center gap-1 sm:gap-4" aria-label="主导航">
            <NavItem icon={<ListChecks size={16} />} label="题库" to="/problems" />
            <NavItem icon={<Send size={16} />} label="提交记录" to="/submissions" hideOnMobile />
            <NavItem icon={<User size={16} />} label="进度" to="/profile" hideOnMobile />
            {!isLoading && (user ? (
              <button
                aria-label="退出登录"
                className="nav-link border-0 bg-transparent"
                onClick={() => void logout()}
                title={`${user.username}，退出登录`}
              >
                <LogOut aria-hidden="true" size={16} />
                <span className="hidden sm:inline">退出</span>
              </button>
            ) : (
              <NavItem icon={<LogIn size={16} />} label="登录" to="/login" />
            ))}
          </nav>
        </div>
      </header>
      <main className={`w-full flex-1 ${isProblemWorkspace ? 'px-0 py-0' : 'px-4 py-6 lg:py-8'}`}><Outlet /></main>
      {!isProblemWorkspace && (
        <footer className="border-t border-[var(--border)] px-4 py-5 text-center text-xs text-[var(--text-muted)]">
          独立学习项目，非 LeetCode 官方产品
        </footer>
      )}
    </div>
  );
}

function NavItem({
  to,
  icon,
  label,
  hideOnMobile = false,
}: {
  to: string;
  icon: React.ReactNode;
  label: string;
  hideOnMobile?: boolean;
}) {
  return (
    <NavLink
      className={({ isActive }) => `nav-link ${isActive ? 'nav-link-active' : ''} ${hideOnMobile ? 'hidden sm:inline-flex' : ''}`}
      to={to}
    >
      {icon}
      <span>{label}</span>
    </NavLink>
  );
}
