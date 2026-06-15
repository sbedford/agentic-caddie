import { useState, useEffect } from 'react'
import { useJson, useVocabulary, Status } from './hooks'

type Player = { id: number; name: string; handicap: number }

type Club = {
  id: number
  player_id: number
  club_name: string
  carry_avg: number | null
  carry_reliable: number | null
  carry_max: number | null
  dispersion_avg_m: number | null
  dispersion_bias: string | null
  sample_size: number
}

// ── Players list ──────────────────────────────────────────────────────────────

function PlayersList({ onSelect }: { onSelect: (id: number) => void }) {
  const { data, loading, error } = useJson<Player[]>('/players/')
  return (
    <>
      <Status loading={loading} error={error} />
      {data && data.length === 0 && <p className="placeholder">No players found.</p>}
      {data && data.length > 0 && (
        <div className="player-grid">
          {data.map(p => (
            <div key={p.id} className="player-card" onClick={() => onSelect(p.id)}>
              <span className="player-card-name">{p.name}</span>
              <div className="player-card-hcp">
                <span className="player-card-handicap">{p.handicap ?? '—'}</span>
                <span className="player-card-hcp-label">HCP</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  )
}

// ── Handicap editor ───────────────────────────────────────────────────────────

function HandicapEditor({
  playerId,
  current,
  onSaved,
}: {
  playerId: number
  current: number
  onSaved: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    setValue(String(current))
    setEditing(true)
    setError(null)
  }

  async function save() {
    const h = parseFloat(value)
    if (isNaN(h)) { setError('Enter a valid number'); return }
    setSaving(true)
    try {
      const r = await fetch(`/players/${playerId}/handicap`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ handicap: h }),
      })
      if (!r.ok) throw new Error(await r.text())
      setEditing(false)
      onSaved()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  if (!editing) {
    return (
      <div className="inline-field">
        <span className="field-label">Handicap</span>
        <span className="field-value">{current}</span>
        <button className="btn-sm" onClick={startEdit}>Edit</button>
      </div>
    )
  }

  return (
    <div className="inline-field">
      <span className="field-label">Handicap</span>
      <input
        className="inline-input"
        type="number"
        step="0.1"
        value={value}
        onChange={e => setValue(e.target.value)}
        onKeyDown={e => { if (e.key === 'Enter') save(); if (e.key === 'Escape') setEditing(false) }}
        autoFocus
      />
      <button className="btn-sm btn-primary" onClick={save} disabled={saving}>
        {saving ? 'Saving…' : 'Save'}
      </button>
      <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>
        Cancel
      </button>
      {error && <span className="inline-error">{error}</span>}
    </div>
  )
}

// ── Club table shared types ───────────────────────────────────────────────────

type ClubDraft = {
  carry_avg: string
  carry_reliable: string
  carry_max: string
  dispersion_avg_m: string
  dispersion_bias: string
  sample_size: string
}

function toDraft(c: Club): ClubDraft {
  return {
    carry_avg:        c.carry_avg        != null ? String(c.carry_avg)        : '',
    carry_reliable:   c.carry_reliable   != null ? String(c.carry_reliable)   : '',
    carry_max:        c.carry_max        != null ? String(c.carry_max)        : '',
    dispersion_avg_m: c.dispersion_avg_m != null ? String(c.dispersion_avg_m) : '',
    dispersion_bias:  c.dispersion_bias  ?? '',
    sample_size:      String(c.sample_size),
  }
}

function emptyDraft(): ClubDraft {
  return { carry_avg: '', carry_reliable: '', carry_max: '', dispersion_avg_m: '', dispersion_bias: '', sample_size: '0' }
}

function fmt(v: number | null) { return v != null ? v : '—' }

function draftToBody(draft: ClubDraft): Record<string, unknown> {
  const body: Record<string, unknown> = { sample_size: parseInt(draft.sample_size) || 0 }
  if (draft.carry_avg !== '')        body.carry_avg        = parseFloat(draft.carry_avg)
  if (draft.carry_reliable !== '')   body.carry_reliable   = parseFloat(draft.carry_reliable)
  if (draft.carry_max !== '')        body.carry_max        = parseFloat(draft.carry_max)
  if (draft.dispersion_avg_m !== '') body.dispersion_avg_m = parseFloat(draft.dispersion_avg_m)
  if (draft.dispersion_bias !== '')  body.dispersion_bias  = draft.dispersion_bias
  return body
}

function DraftInputs({ draft, onChange }: { draft: ClubDraft; onChange: (key: keyof ClubDraft) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void }) {
  const { vocab }   = useVocabulary()
  const biasOptions = vocab['dispersion_bias'] ?? []

  return (
    <>
      <td><input className="cell-input" type="number" value={draft.carry_avg}        onChange={onChange('carry_avg')}        placeholder="—" /></td>
      <td><input className="cell-input" type="number" value={draft.carry_reliable}   onChange={onChange('carry_reliable')}   placeholder="—" /></td>
      <td><input className="cell-input" type="number" value={draft.carry_max}        onChange={onChange('carry_max')}        placeholder="—" /></td>
      <td><input className="cell-input" type="number" value={draft.dispersion_avg_m} onChange={onChange('dispersion_avg_m')} placeholder="—" /></td>
      <td>
        <select className="cell-input" value={draft.dispersion_bias} onChange={onChange('dispersion_bias')}>
          <option value="">—</option>
          {biasOptions.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </td>
      <td><input className="cell-input" type="number" value={draft.sample_size}      onChange={onChange('sample_size')} /></td>
    </>
  )
}

// ── Existing club row (view / edit / retire) ──────────────────────────────────

function ClubRow({
  club,
  playerId,
  editing,
  onEdit,
  onCancel,
  onSaved,
  onRetired,
}: {
  club: Club
  playerId: number
  editing: boolean
  onEdit: () => void
  onCancel: () => void
  onSaved: () => void
  onRetired: () => void
}) {
  const [draft, setDraft] = useState<ClubDraft>(toDraft(club))
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [confirming, setConfirming] = useState(false)
  const [retiring, setRetiring] = useState(false)

  useEffect(() => {
    if (editing) { setDraft(toDraft(club)); setError(null); setConfirming(false) }
  }, [editing]) // eslint-disable-line react-hooks/exhaustive-deps

  function field(key: keyof ClubDraft) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    setSaving(true)
    setError(null)
    try {
      const r = await fetch(`/clubs/${club.id}/distances`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(draftToBody(draft)),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  async function retire() {
    setRetiring(true)
    try {
      const today = new Date().toISOString().slice(0, 10)
      const r = await fetch(`/clubs/player/${playerId}/name/${encodeURIComponent(club.club_name)}/retire`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ removed_date: today }),
      })
      if (!r.ok) throw new Error(await r.text())
      onRetired()
    } catch (e: any) {
      setError(e.message)
      setRetiring(false)
      setConfirming(false)
    }
  }

  if (!editing) {
    return (
      <tr>
        <td>{club.club_name}</td>
        <td>{fmt(club.carry_avg)}</td>
        <td>{fmt(club.carry_reliable)}</td>
        <td>{fmt(club.carry_max)}</td>
        <td>{fmt(club.dispersion_avg_m)}</td>
        <td>{club.dispersion_bias ?? '—'}</td>
        <td>{club.sample_size}</td>
        <td>
          {!confirming ? (
            <>
              <button className="btn-sm" onClick={onEdit}>Edit</button>
              {' '}
              <button className="btn-sm btn-danger" onClick={() => setConfirming(true)}>Retire</button>
            </>
          ) : (
            <>
              <span className="confirm-label">Remove from bag?</span>
              <button className="btn-sm btn-danger" onClick={retire} disabled={retiring}>
                {retiring ? '…' : 'Yes'}
              </button>
              {' '}
              <button className="btn-sm" onClick={() => setConfirming(false)}>No</button>
            </>
          )}
          {error && <div className="inline-error" style={{ marginTop: 4 }}>{error}</div>}
        </td>
      </tr>
    )
  }

  return (
    <tr className="row-editing">
      <td><strong>{club.club_name}</strong></td>
      <DraftInputs draft={draft} onChange={field} />
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>
          {saving ? '…' : 'Save'}
        </button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error" style={{ marginTop: 4 }}>{error}</div>}
      </td>
    </tr>
  )
}

// ── Add club row ──────────────────────────────────────────────────────────────

function AddClubRow({ playerId, onCancel, onSaved }: { playerId: number; onCancel: () => void; onSaved: () => void }) {
  const [name, setName] = useState('')
  const [draft, setDraft] = useState<ClubDraft>(emptyDraft())
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function field(key: keyof ClubDraft) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    if (!name.trim()) { setError('Club name is required'); return }
    setSaving(true)
    setError(null)
    try {
      const today = new Date().toISOString().slice(0, 10)
      const r = await fetch('/clubs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          player_id: playerId,
          club_name: name.trim(),
          added_date: today,
          ...draftToBody(draft),
        }),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) {
      setError(e.message)
    } finally {
      setSaving(false)
    }
  }

  return (
    <tr className="row-adding">
      <td>
        <input
          className="cell-input"
          type="text"
          value={name}
          onChange={e => setName(e.target.value)}
          placeholder="e.g. 7i"
          style={{ width: 80 }}
          autoFocus
        />
      </td>
      <DraftInputs draft={draft} onChange={field} />
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>
          {saving ? '…' : 'Add'}
        </button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error" style={{ marginTop: 4 }}>{error}</div>}
      </td>
    </tr>
  )
}

// ── Player profile ────────────────────────────────────────────────────────────

function PlayerProfile({ id, onBack }: { id: number; onBack: () => void }) {
  const [playerBust, setPlayerBust] = useState(0)
  const [clubsBust,  setClubsBust]  = useState(0)
  const [editingClubId, setEditingClubId] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)

  const { data: player, loading: pLoading, error: pError } = useJson<Player>(`/players/${id}`, playerBust)
  const { data: clubs,  loading: cLoading, error: cError  } = useJson<Club[]>(`/clubs/player/${id}/active`, clubsBust)

  const sorted = clubs
    ? [...clubs].sort((a, b) => (b.carry_reliable ?? -Infinity) - (a.carry_reliable ?? -Infinity))
    : null

  function refreshClubs() {
    setEditingClubId(null)
    setAdding(false)
    setClubsBust(b => b + 1)
  }

  function startAdding() {
    setEditingClubId(null)
    setAdding(true)
  }

  return (
    <div>
      <button className="btn-back" onClick={onBack}>← Players</button>

      {pLoading && <p className="status">Loading…</p>}
      {pError   && <p className="error">{pError}</p>}
      {player && (
        <>
          <h1>{player.name}</h1>
          <HandicapEditor
            playerId={id}
            current={player.handicap}
            onSaved={() => setPlayerBust(b => b + 1)}
          />
        </>
      )}

      <div className="section-header">
        <h2>Clubs in bag</h2>
        {!adding && (
          <button className="btn-sm btn-primary" onClick={startAdding}>+ Add club</button>
        )}
      </div>

      <Status loading={cLoading} error={cError} />

      {sorted && (sorted.length > 0 || adding) && (
        <table>
          <thead>
            <tr>
              <th>Club</th>
              <th>Carry avg (m)</th>
              <th>Reliable (m)</th>
              <th>Max (m)</th>
              <th>Dispersion (m)</th>
              <th>Direction</th>
              <th>n</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {sorted.map(c => (
              <ClubRow
                key={c.id}
                club={c}
                playerId={id}
                editing={editingClubId === c.id}
                onEdit={() => { setAdding(false); setEditingClubId(c.id) }}
                onCancel={() => setEditingClubId(null)}
                onSaved={refreshClubs}
                onRetired={refreshClubs}
              />
            ))}
            {adding && (
              <AddClubRow
                playerId={id}
                onCancel={() => setAdding(false)}
                onSaved={refreshClubs}
              />
            )}
          </tbody>
        </table>
      )}

      {sorted && sorted.length === 0 && !adding && (
        <p className="placeholder">No active clubs recorded.</p>
      )}
    </div>
  )
}

// ── Section root ──────────────────────────────────────────────────────────────

export function PlayersSection({ initialId }: { initialId?: number }) {
  const [selectedId, setSelectedId] = useState<number | null>(initialId ?? null)

  if (selectedId !== null) {
    return <PlayerProfile id={selectedId} onBack={() => setSelectedId(null)} />
  }

  return (
    <>
      <h1>Players</h1>
      <PlayersList onSelect={setSelectedId} />
    </>
  )
}
