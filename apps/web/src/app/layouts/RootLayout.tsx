import { Link, Outlet } from 'react-router';
import { Flame, ListChecks, Send, User } from 'lucide-react';

export default function RootLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <header
        className="sticky top-0 z-50 border-b"
        style={{
          backgroundColor: 'var(--surface)',
          borderColor: 'var(--border)',
        }}
      >
        <div className="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 font-bold text-lg no-underline" style={{ color: 'var(--accent)' }}>
            <Flame size={24} />
            ACM HOT 100
          </Link>
          <nav className="flex items-center gap-6">
            <Link
              to="/problems"
              className="flex items-center gap-1.5 text-sm font-medium no-underline transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--text)')}
              onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
            >
              <ListChecks size={16} />
              Problems
            </Link>
            <Link
              to="/submissions"
              className="flex items-center gap-1.5 text-sm font-medium no-underline transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--text)')}
              onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
            >
              <Send size={16} />
              Submissions
            </Link>
            <Link
              to="/profile"
              className="flex items-center gap-1.5 text-sm font-medium no-underline transition-colors"
              style={{ color: 'var(--text-muted)' }}
              onMouseEnter={(e) => (e.currentTarget.style.color = 'var(--text)')}
              onMouseLeave={(e) => (e.currentTarget.style.color = 'var(--text-muted)')}
            >
              <User size={16} />
              Profile
            </Link>
          </nav>
        </div>
      </header>
      <main className="flex-1 max-w-7xl mx-auto w-full px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
