export default function HomePage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[60vh]">
      <h1 className="text-4xl font-bold mb-4" style={{ color: 'var(--accent)' }}>
        ACM HOT 100
      </h1>
      <p className="text-lg" style={{ color: 'var(--text-muted)' }}>
        Master the top 100 algorithm problems
      </p>
      <a
        href="/problems"
        className="mt-8 px-6 py-3 rounded-lg text-white font-medium transition-colors"
        style={{ backgroundColor: 'var(--accent)' }}
        onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = 'var(--accent-hover)')}
        onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'var(--accent)')}
      >
        Start Practicing
      </a>
    </div>
  );
}
