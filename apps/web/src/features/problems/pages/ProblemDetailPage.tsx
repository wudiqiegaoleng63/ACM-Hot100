import { useParams } from 'react-router';

export default function ProblemDetailPage() {
  const { slug } = useParams<{ slug: string }>();

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">
        Problem: {slug}
      </h1>
      <p style={{ color: 'var(--text-muted)' }}>
        Problem detail page coming soon
      </p>
    </div>
  );
}
