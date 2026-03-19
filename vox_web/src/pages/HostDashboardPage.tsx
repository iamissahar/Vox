import React, { useEffect, useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { hubApi, userApi } from '../api';
import { useAuth } from '../hooks/useAuth';
import { VoxLogo } from '../components/VoxLogo';
import { Button } from '../components/Button';
import { Card } from '../components/Card';

// ─── Hub card ─────────────────────────────────────────────────────────────────

function HubCard({
  hubId,
  userId,
  onDeleted,
}: {
  hubId: string;
  userId: string;
  onDeleted: () => void;
}) {
  const navigate = useNavigate();
  const [isDeleting, setIsDeleting] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleDelete = async () => {
    if (!window.confirm('Delete this hub? All listeners will be disconnected.')) return;
    setIsDeleting(true);
    try {
      await hubApi.delete(hubId, userId);
      onDeleted();
    } catch {
      alert('Failed to delete hub. Please try again.');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(hubId);
    setCopied(true);
    setTimeout(() => setCopied(false), 1800);
  };

  return (
    <Card>
      {/* Hub ID badge */}
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          marginBottom: '20px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
          <div
            style={{
              width: '10px',
              height: '10px',
              borderRadius: '50%',
              background: 'var(--success)',
              boxShadow: '0 0 8px var(--success)',
            }}
          />
          <span style={{ fontSize: '12px', color: 'var(--text-muted)', fontWeight: 500 }}>
            HUB
          </span>
        </div>
        <button
          onClick={handleDelete}
          disabled={isDeleting}
          style={{
            background: 'none',
            border: 'none',
            cursor: 'pointer',
            color: 'var(--text-muted)',
            fontSize: '18px',
            lineHeight: 1,
            transition: 'color 0.2s',
          }}
          onMouseEnter={(e) => ((e.target as HTMLElement).style.color = 'var(--error)')}
          onMouseLeave={(e) => ((e.target as HTMLElement).style.color = 'var(--text-muted)')}
          title="Delete hub"
        >
          {isDeleting ? '…' : '×'}
        </button>
      </div>

      {/* ID display */}
      <div
        style={{
          background: 'var(--bg-elevated)',
          border: '1px solid var(--border)',
          borderRadius: '8px',
          padding: '10px 14px',
          fontFamily: 'monospace',
          fontSize: '13px',
          color: 'var(--text-secondary)',
          marginBottom: '20px',
          wordBreak: 'break-all',
        }}
      >
        {hubId}
      </div>

      {/* Actions */}
      <div style={{ display: 'flex', gap: '10px', flexWrap: 'wrap' }}>
        <Button
          size="sm"
          onClick={() => navigate(`/host/${hubId}`)}
          style={{ flex: 1 }}
        >
          Go Live →
        </Button>
        <Button
          size="sm"
          variant="secondary"
          onClick={handleCopy}
          style={{ flex: 1 }}
        >
          {copied ? '✓ Copied' : 'Copy ID'}
        </Button>
      </div>
    </Card>
  );
}

// ─── Main ─────────────────────────────────────────────────────────────────────

export function HostDashboardPage() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const [hubIds, setHubIds] = useState<string[]>([]);
  const [isLoadingHubs, setIsLoadingHubs] = useState(true);
  const [isCreating, setIsCreating] = useState(false);

  const loadHubs = useCallback(async () => {
    setIsLoadingHubs(true);
    try {
      const { data } = await userApi.getHubs();
      setHubIds(data.hub_ids || []);
    } catch {
      setHubIds([]);
    } finally {
      setIsLoadingHubs(false);
    }
  }, []);

  useEffect(() => {
    loadHubs();
  }, [loadHubs]);

  const handleCreateHub = async () => {
    setIsCreating(true);
    try {
      const { data } = await hubApi.create();
      setHubIds((prev) => [...prev, data.hub_id]);
    } catch {
      alert('Failed to create hub. Please try again.');
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Topbar */}
      <nav
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '16px 40px',
          borderBottom: '1px solid var(--border)',
        }}
      >
        <VoxLogo size={28} />

        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <button
            onClick={() => navigate('/profile')}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '10px',
              background: 'none',
              border: 'none',
              cursor: 'pointer',
              color: 'var(--text-secondary)',
              fontSize: '14px',
            }}
          >
            {user?.picture ? (
              <img
                src={user.picture}
                alt={user.name}
                style={{ width: '30px', height: '30px', borderRadius: '50%', objectFit: 'cover' }}
              />
            ) : (
              <div
                style={{
                  width: '30px',
                  height: '30px',
                  borderRadius: '50%',
                  background: 'var(--accent-dim)',
                  border: '1px solid var(--border-accent)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: '13px',
                  color: 'var(--accent)',
                  fontWeight: 600,
                }}
              >
                {user?.name?.[0]?.toUpperCase() ?? '?'}
              </div>
            )}
            <span>{user?.name ?? user?.email}</span>
          </button>
          <Button variant="ghost" size="sm" onClick={logout}>
            Sign out
          </Button>
        </div>
      </nav>

      {/* Content */}
      <main
        style={{
          flex: 1,
          maxWidth: '900px',
          margin: '0 auto',
          padding: '48px 24px',
          width: '100%',
        }}
      >
        {/* Header row */}
        <div
          style={{
            display: 'flex',
            alignItems: 'flex-end',
            justifyContent: 'space-between',
            marginBottom: '36px',
            flexWrap: 'wrap',
            gap: '16px',
          }}
        >
          <div>
            <h1 style={{ fontSize: '32px', marginBottom: '6px' }}>Your Hubs</h1>
            <p style={{ color: 'var(--text-secondary)', fontSize: '15px' }}>
              Create a hub and share its ID with your listeners.
            </p>
          </div>
          <Button onClick={handleCreateHub} isLoading={isCreating} size="md">
            + New Hub
          </Button>
        </div>

        {/* Grid */}
        {isLoadingHubs ? (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              padding: '80px',
              color: 'var(--text-muted)',
            }}
          >
            Loading hubs…
          </div>
        ) : hubIds.length === 0 ? (
          <div
            style={{
              textAlign: 'center',
              padding: '80px 24px',
              color: 'var(--text-muted)',
            }}
          >
            <div style={{ fontSize: '40px', marginBottom: '16px' }}>🎙</div>
            <p style={{ marginBottom: '8px', fontSize: '16px', color: 'var(--text-secondary)' }}>
              No hubs yet
            </p>
            <p style={{ fontSize: '14px' }}>Create your first hub to start broadcasting.</p>
          </div>
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
              gap: '20px',
            }}
          >
            {hubIds.map((id) => (
              <HubCard
                key={id}
                hubId={id}
                userId={user!.id}
                onDeleted={() => setHubIds((prev) => prev.filter((h) => h !== id))}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
