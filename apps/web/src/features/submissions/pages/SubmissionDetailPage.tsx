import { useParams } from 'react-router';

export default function SubmissionDetailPage() {
  const { id } = useParams<{ id: string }>();

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">
        Submission: {id}
      </h1>
      <p style={{ color: 'var(--text-muted)' }}>
        Submission detail page coming soon
      </p>
    </div>
  );
}
