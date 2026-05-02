import { useEffect, useState } from 'react'

type DailyCounts = Record<string, number>

function TokenForm({ onSubmit }: { onSubmit: (token: string) => void }) {
  const [input, setInput] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (input.trim()) onSubmit(input.trim())
  }

  return (
    <div style={styles.centered}>
      <h2 style={styles.heading}>CryptoPulse Admin</h2>
      <form onSubmit={handleSubmit} style={styles.form}>
        <input
          type="password"
          placeholder="Enter stats token"
          value={input}
          onChange={e => setInput(e.target.value)}
          style={styles.input}
          autoFocus
        />
        <button type="submit" style={styles.button}>View Stats</button>
      </form>
    </div>
  )
}

export default function AdminPage() {
  const [token, setToken] = useState<string>(sessionStorage.getItem('stats_token') ?? '')
  const [counts, setCounts] = useState<DailyCounts | null>(null)
  const [error, setError] = useState('')

  const fetchStats = async (t: string) => {
    setError('')
    try {
      const res = await fetch('/stats', {
        headers: t ? { Authorization: `Bearer ${t}` } : {},
      })
      if (res.status === 401) {
        setError('Invalid token.')
        sessionStorage.removeItem('stats_token')
        setToken('')
        return
      }
      if (!res.ok) {
        setError(`Error: ${res.status} ${res.statusText}`)
        return
      }
      const data: DailyCounts = await res.json()
      setCounts(data)
    } catch {
      setError('Could not reach the server.')
    }
  }

  useEffect(() => {
    if (token) fetchStats(token)
  }, [])

  const handleTokenSubmit = (t: string) => {
    sessionStorage.setItem('stats_token', t)
    setToken(t)
    fetchStats(t)
  }

  if (!token) return <TokenForm onSubmit={handleTokenSubmit} />

  const sortedDates = counts
    ? Object.keys(counts).sort((a, b) => b.localeCompare(a))
    : []

  const totalVisitors = counts ? Object.values(counts).reduce((s, v) => s + v, 0) : 0

  return (
    <div style={styles.centered}>
      <h2 style={styles.heading}>CryptoPulse — Daily Visitors</h2>
      {error && <p style={styles.error}>{error}</p>}
      {counts && (
        <>
          <p style={styles.summary}>Total tracked visitors: <strong>{totalVisitors}</strong></p>
          <table style={styles.table}>
            <thead>
              <tr>
                <th style={styles.th}>Date (UTC)</th>
                <th style={styles.th}>Unique Visitors</th>
              </tr>
            </thead>
            <tbody>
              {sortedDates.map(date => (
                <tr key={date}>
                  <td style={styles.td}>{date}</td>
                  <td style={{ ...styles.td, textAlign: 'center' }}>{counts[date]}</td>
                </tr>
              ))}
              {sortedDates.length === 0 && (
                <tr>
                  <td colSpan={2} style={{ ...styles.td, textAlign: 'center', color: '#888' }}>No data yet.</td>
                </tr>
              )}
            </tbody>
          </table>
          <button
            onClick={() => { sessionStorage.removeItem('stats_token'); setToken(''); setCounts(null) }}
            style={{ ...styles.button, marginTop: '1.5rem', background: '#555' }}
          >
            Sign out
          </button>
        </>
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  centered: {
    minHeight: '100vh',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    fontFamily: 'monospace',
    background: '#0f0f0f',
    color: '#e0e0e0',
    padding: '2rem',
  },
  heading: {
    fontSize: '1.4rem',
    marginBottom: '1.5rem',
    letterSpacing: '0.05em',
  },
  form: {
    display: 'flex',
    flexDirection: 'column',
    gap: '0.75rem',
    width: '280px',
  },
  input: {
    padding: '0.6rem 0.8rem',
    fontSize: '1rem',
    borderRadius: '4px',
    border: '1px solid #444',
    background: '#1a1a1a',
    color: '#e0e0e0',
    outline: 'none',
  },
  button: {
    padding: '0.6rem',
    fontSize: '0.9rem',
    borderRadius: '4px',
    border: 'none',
    background: '#f3ba2f',
    color: '#000',
    cursor: 'pointer',
    fontWeight: 600,
  },
  summary: {
    marginBottom: '1rem',
    fontSize: '0.9rem',
    color: '#aaa',
  },
  table: {
    borderCollapse: 'collapse',
    width: '360px',
  },
  th: {
    padding: '0.6rem 1rem',
    borderBottom: '1px solid #333',
    textAlign: 'left',
    fontSize: '0.85rem',
    color: '#f3ba2f',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  },
  td: {
    padding: '0.55rem 1rem',
    borderBottom: '1px solid #222',
    fontSize: '0.9rem',
  },
  error: {
    color: '#ff6b6b',
    marginBottom: '1rem',
    fontSize: '0.9rem',
  },
}
