import { useState, useEffect, useRef } from 'react'
import { useJson, useVocabulary, Status } from './hooks'
import type { NavTarget } from './types'

// ── Types ─────────────────────────────────────────────────────────────────────

type Player = { id: number; name: string }
type Course = { id: number; name: string }
type Round = {
  id: number; player_id: number; course_id: number
  played_at: string; tees: string; round_type: string
  total_score: number | null; total_points: number | null; total_putts: number | null
}
type Hole = {
  id: number; round_id: number; course_hole_id: number; hole_number: number
  flag_position: string | null; score: number | null; points: number | null
  putts: number | null; gir: boolean | null; scramble_save: boolean | null; penalty: boolean | null
}
type Shot = {
  id: number; hole_id: number; shot_number: number; shot_type: string
  club: string | null; result: string | null; miss: string | null
  strike_quality: string | null; source: string
}
type Club    = { id: number; club_name: string; carry_reliable: number | null }
type Tee     = { id: number; course_id: number; name: string }
type TeeHole = { id: number; course_hole_id: number; tee_id: number; par: number }
type NameMap = Record<number, string>
type ParMap  = Record<number, number>  // course_hole_id → par

// ── Constants ─────────────────────────────────────────────────────────────────

const NO_MISS_RESULTS = new Set(['fairway', 'green'])

// ── Helpers ───────────────────────────────────────────────────────────────────

function fmtDate(iso: string) {
  const [y, m, d] = iso.slice(0, 10).split('-').map(Number)
  return new Date(y, m - 1, d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

function boolLabel(v: boolean | null) {
  if (v === null) return <span className="bool-null">—</span>
  return v ? <span className="bool-yes">✓</span> : <span className="bool-no">✗</span>
}

function sum(holes: Hole[], key: 'score' | 'points' | 'putts'): number {
  return holes.reduce((acc, h) => acc + (h[key] ?? 0), 0)
}

function scoreClass(points: number | null) {
  if (points === null) return ''
  if (points >= 4) return 'pts-eagle'
  if (points === 3) return 'pts-birdie'
  if (points === 2) return 'pts-par'
  if (points === 1) return 'pts-bogey'
  return 'pts-double'
}

// ── BoolToggle ────────────────────────────────────────────────────────────────

function BoolToggle({ value, onChange }: {
  value: boolean | null
  onChange: (v: boolean | null) => void
}) {
  return (
    <span className="bool-toggle">
      <button className={value === null  ? 'active' : ''} onClick={() => onChange(null)}>—</button>
      <button className={value === true  ? 'active' : ''} onClick={() => onChange(true)}>✓</button>
      <button className={value === false ? 'active' : ''} onClick={() => onChange(false)}>✗</button>
    </span>
  )
}

// ── Drafts ────────────────────────────────────────────────────────────────────

type HoleDraft = {
  score: string; points: string; putts: string; flag_position: string
  gir: boolean | null; scramble_save: boolean | null; penalty: boolean | null
}

function holeToDraft(h: Hole): HoleDraft {
  return {
    score:         h.score  != null ? String(h.score)  : '',
    points:        h.points != null ? String(h.points) : '',
    putts:         h.putts  != null ? String(h.putts)  : '',
    flag_position: h.flag_position ?? '',
    gir:           h.gir,
    scramble_save: h.scramble_save,
    penalty:       h.penalty,
  }
}

type ShotDraft = {
  shot_type: string; club: string; result: string; miss: string; strike_quality: string
}

function shotToDraft(s: Shot): ShotDraft {
  return {
    shot_type:      s.shot_type,
    club:           s.club           ?? '',
    result:         s.result         ?? '',
    miss:           s.miss           ?? '',
    strike_quality: s.strike_quality ?? '',
  }
}

// ── Shot edit fields ──────────────────────────────────────────────────────────

function ShotEditFields({ draft, set, clubs, onClearMiss }: {
  draft: ShotDraft
  clubs: Club[]
  set: <K extends keyof ShotDraft>(key: K) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
  onClearMiss: () => void
}) {
  const { vocab }  = useVocabulary()
  const shotTypes  = vocab['shot_type']    ?? []
  const results    = vocab['shot_result']  ?? []
  const misses     = vocab['shot_miss']    ?? []
  const strikes    = vocab['shot_strike']  ?? []
  const showMiss   = !NO_MISS_RESULTS.has(draft.result)

  function handleResultChange(e: React.ChangeEvent<HTMLSelectElement>) {
    set('result')(e)
    if (NO_MISS_RESULTS.has(e.target.value)) onClearMiss()
  }

  return (
    <>
      <div className="shot-edit-row">
        <label className="shot-edit-label">Type</label>
        <select className="cell-input" value={draft.shot_type} onChange={set('shot_type')} style={{ width: 110 }}>
          {shotTypes.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
        <label className="shot-edit-label">Club</label>
        <select className="cell-input" value={draft.club} onChange={set('club')} style={{ width: 110 }}>
          <option value="">—</option>
          {clubs.map(c => <option key={c.id} value={c.club_name}>{c.club_name}</option>)}
        </select>
      </div>
      <div className="shot-edit-row">
        <label className="shot-edit-label">Result</label>
        <select className="cell-input" value={draft.result} onChange={handleResultChange} style={{ width: 110 }}>
          <option value="">—</option>
          {results.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
        {showMiss && (
          <>
            <label className="shot-edit-label">Miss</label>
            <select className="cell-input" value={draft.miss} onChange={set('miss')} style={{ width: 90 }}>
              <option value="">—</option>
              {misses.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
            </select>
          </>
        )}
        <label className="shot-edit-label">Strike</label>
        <select className="cell-input" value={draft.strike_quality} onChange={set('strike_quality')} style={{ width: 90 }}>
          <option value="">—</option>
          {strikes.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </div>
    </>
  )
}

// ── Shot row (full-width card, view + edit) ───────────────────────────────────

function ShotRow({ shot, clubs, onSaved, onDeleted, isDragOver, onDragStart, onDragOver, onDragLeave, onDrop, onDragEnd }: {
  shot: Shot; clubs: Club[]; onSaved: () => void; onDeleted: () => void
  isDragOver: boolean
  onDragStart: () => void
  onDragOver: (e: React.DragEvent) => void
  onDragLeave: () => void
  onDrop: () => void
  onDragEnd: () => void
}) {
  const [editing, setEditing]       = useState(false)
  const [draft, setDraft]           = useState<ShotDraft>(shotToDraft(shot))
  const [saving, setSaving]         = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [deleting, setDeleting]     = useState(false)
  const [error, setError]           = useState<string | null>(null)

  function set<K extends keyof ShotDraft>(key: K) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    setSaving(true); setError(null)
    const body: Record<string, unknown> = { shot_type: draft.shot_type, source: shot.source || 'manual' }
    if (draft.club !== '')           body.club           = draft.club
    if (draft.result !== '')         body.result         = draft.result
    if (draft.miss !== '')           body.miss           = draft.miss
    if (draft.strike_quality !== '') body.strike_quality = draft.strike_quality
    try {
      const r = await fetch(`/shots/${shot.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
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

  async function doDelete() {
    setDeleting(true)
    try {
      const r = await fetch(`/shots/${shot.id}`, { method: 'DELETE' })
      if (!r.ok) throw new Error(await r.text())
      onDeleted()
    } catch (e: any) {
      setError(e.message)
      setDeleting(false)
      setConfirming(false)
    }
  }

  const meta = [shot.club, shot.result, shot.miss, shot.strike_quality].filter(Boolean).join(' · ')

  if (!editing) {
    return (
      <div
        className={`shot-row${isDragOver ? ' shot-row-drag-over' : ''}`}
        data-shot-type={shot.shot_type}
        draggable
        onDragStart={onDragStart}
        onDragOver={onDragOver}
        onDragLeave={onDragLeave}
        onDrop={onDrop}
        onDragEnd={onDragEnd}
      >
        <div className={`shot-row-badge shot-type-${shot.shot_type}`}>{shot.shot_number}</div>
        <div className="shot-row-body">
          <div className="shot-row-type">{shot.shot_type}</div>
          {meta && <div className="shot-row-meta">{meta}</div>}
        </div>
        <div className="shot-row-actions">
          <span className="shot-row-drag">⠿</span>
          <button className="btn-sm" onClick={() => { setDraft(shotToDraft(shot)); setEditing(true) }}>Edit</button>
        </div>
      </div>
    )
  }

  return (
    <div className="shot-row shot-row-editing" data-shot-type={shot.shot_type}>
      <div className="shot-row-edit-header">
        <div className={`shot-row-badge shot-type-${shot.shot_type}`}>{shot.shot_number}</div>
        <span className="shot-row-edit-title">Shot {shot.shot_number}</span>
        <button className="btn-sm" onClick={() => { setDraft(shotToDraft(shot)); setEditing(false) }}>✕</button>
      </div>
      <ShotEditFields draft={draft} set={set} clubs={clubs}
        onClearMiss={() => setDraft(d => ({ ...d, miss: '' }))} />
      <div className="form-actions">
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>Cancel</button>
        {!confirming
          ? <button className="btn-sm btn-danger" onClick={() => setConfirming(true)} disabled={saving}>Delete</button>
          : <>
              <span className="confirm-label">Delete?</span>
              <button className="btn-sm btn-danger" onClick={doDelete} disabled={deleting}>{deleting ? '…' : 'Yes'}</button>
              <button className="btn-sm" onClick={() => setConfirming(false)}>No</button>
            </>
        }
        {error && <span className="inline-error">{error}</span>}
      </div>
    </div>
  )
}

// ── Add shot row ──────────────────────────────────────────────────────────────

function AddShotRow({ holeId, nextShotNum, clubs, onSaved, onCancel }: {
  holeId: number; nextShotNum: number; clubs: Club[]; onSaved: () => void; onCancel: () => void
}) {
  const [draft, setDraft] = useState<ShotDraft>({
    shot_type: 'tee', club: '', result: '', miss: '', strike_quality: '',
  })
  const [saving, setSaving] = useState(false)
  const [error, setError]   = useState<string | null>(null)

  function set<K extends keyof ShotDraft>(key: K) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) =>
      setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    setSaving(true); setError(null)
    const body: Record<string, unknown> = {
      hole_id: holeId, shot_number: nextShotNum, shot_type: draft.shot_type, source: 'manual',
    }
    if (draft.club !== '')           body.club           = draft.club
    if (draft.result !== '')         body.result         = draft.result
    if (draft.miss !== '')           body.miss           = draft.miss
    if (draft.strike_quality !== '') body.strike_quality = draft.strike_quality
    try {
      const r = await fetch('/shots', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
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
    <div className="shot-row shot-row-editing" data-shot-type={draft.shot_type}>
      <div className="shot-row-edit-header">
        <div className={`shot-row-badge shot-type-${draft.shot_type}`}>{nextShotNum}</div>
        <span className="shot-row-edit-title">Shot {nextShotNum}</span>
      </div>
      <ShotEditFields draft={draft} set={set} clubs={clubs}
        onClearMiss={() => setDraft(d => ({ ...d, miss: '' }))} />
      <div className="form-actions">
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Add shot'}</button>
        <button className="btn-sm" onClick={onCancel} disabled={saving}>Cancel</button>
        {error && <span className="inline-error">{error}</span>}
      </div>
    </div>
  )
}

// ── Hole stats strip (view) ───────────────────────────────────────────────────

function HoleStatsStrip({ hole, onEdit }: { hole: Hole; onEdit: () => void }) {
  return (
    <div className="hole-stats-strip">
      <div className="hss-item">
        <span className="hss-val">{hole.points ?? '—'}</span>
        <span className="hss-lbl">Points</span>
      </div>
      <div className="hss-item">
        <span className="hss-val">{hole.putts ?? '—'}</span>
        <span className="hss-lbl">Putts</span>
      </div>
      <div className="hss-sep" />
      <div className="hss-item">
        {boolLabel(hole.gir)}
        <span className="hss-lbl">GIR</span>
      </div>
      <div className="hss-item">
        {boolLabel(hole.scramble_save)}
        <span className="hss-lbl">Scramble</span>
      </div>
      <div className="hss-item">
        {boolLabel(hole.penalty)}
        <span className="hss-lbl">Penalty</span>
      </div>
      {hole.flag_position && (
        <>
          <div className="hss-sep" />
          <div className="hss-item">
            <span className="hss-val hss-val-sm">{hole.flag_position.replace('_', ' ')}</span>
            <span className="hss-lbl">Flag</span>
          </div>
        </>
      )}
      <button className="btn-sm hss-edit-btn" onClick={onEdit}>Edit</button>
    </div>
  )
}

// ── Hole stats form (edit) ────────────────────────────────────────────────────

function HoleStatsForm({ hole, onSaved, onCancel }: {
  hole: Hole; onSaved: () => void; onCancel: () => void
}) {
  const { vocab }      = useVocabulary()
  const flagPositions  = vocab['flag_position'] ?? []
  const [draft, setDraft] = useState<HoleDraft>(holeToDraft(hole))
  const [saving, setSaving] = useState(false)
  const [error, setError]   = useState<string | null>(null)

  function set<K extends keyof HoleDraft>(key: K, val: HoleDraft[K]) {
    setDraft(d => ({ ...d, [key]: val }))
  }

  async function save() {
    setSaving(true); setError(null)
    const body: Record<string, unknown> = {}
    if (draft.score !== '')           body.score         = parseInt(draft.score)
    if (draft.points !== '')          body.points        = parseInt(draft.points)
    if (draft.putts !== '')           body.putts         = parseInt(draft.putts)
    if (draft.flag_position !== '')   body.flag_position = draft.flag_position
    if (draft.gir !== null)           body.gir           = draft.gir
    if (draft.scramble_save !== null) body.scramble_save = draft.scramble_save
    if (draft.penalty !== null)       body.penalty       = draft.penalty
    try {
      const r = await fetch(`/holes/${hole.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
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
    <div className="hole-stats-form-card">
      <div className="hole-stats-form-row">
        <label className="hole-stats-label">Score</label>
        <input className="cell-input" type="number" value={draft.score}
          onChange={e => set('score', e.target.value)} placeholder="—" style={{ width: 64 }} />
        <label className="hole-stats-label">Points</label>
        <input className="cell-input" type="number" value={draft.points}
          onChange={e => set('points', e.target.value)} placeholder="—" style={{ width: 64 }} />
        <label className="hole-stats-label">Putts</label>
        <input className="cell-input" type="number" value={draft.putts}
          onChange={e => set('putts', e.target.value)} placeholder="—" style={{ width: 64 }} />
        <label className="hole-stats-label">Flag</label>
        <select className="cell-input" value={draft.flag_position}
          onChange={e => set('flag_position', e.target.value)} style={{ width: 140 }}>
          <option value="">—</option>
          {flagPositions.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </div>
      <div className="hole-stats-form-row">
        <label className="hole-stats-label">GIR</label>
        <BoolToggle value={draft.gir}           onChange={v => set('gir', v)} />
        <label className="hole-stats-label" style={{ marginLeft: 16 }}>Scramble</label>
        <BoolToggle value={draft.scramble_save} onChange={v => set('scramble_save', v)} />
        <label className="hole-stats-label" style={{ marginLeft: 16 }}>Penalty</label>
        <BoolToggle value={draft.penalty}       onChange={v => set('penalty', v)} />
      </div>
      <div className="form-actions">
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        <button className="btn-sm" onClick={onCancel} disabled={saving}>Cancel</button>
        {error && <span className="inline-error">{error}</span>}
      </div>
    </div>
  )
}

// ── Hole detail page ──────────────────────────────────────────────────────────

function HoleDetail({ hole, clubs, onBack, onSaved, onPrev, onNext }: {
  hole: Hole; clubs: Club[]; onBack: () => void; onSaved: () => void
  onPrev?: () => void; onNext?: () => void
}) {
  const [holeBust, setHoleBust]         = useState(0)
  const { data: freshHole }             = useJson<Hole>(`/holes/${hole.id}`, holeBust)
  const current                          = freshHole ?? hole
  const [editingStats, setEditingStats] = useState(false)

  useEffect(() => {
    setEditingStats(false)
    setAdding(false)
  }, [hole.id]) // eslint-disable-line react-hooks/exhaustive-deps

  const [shotsBust, setShotsBust]       = useState(0)
  const [shots, setShots]               = useState<Shot[] | null>(null)
  const [shotsLoading, setShotsLoading] = useState(true)
  const [localShots, setLocalShots]     = useState<Shot[]>([])
  const [adding, setAdding]             = useState(false)
  const [reorderErr, setReorderErr]     = useState<string | null>(null)
  const draggedId                        = useRef<number | null>(null)
  const [dragOverId, setDragOverId]     = useState<number | null>(null)

  useEffect(() => {
    if (shots) setLocalShots([...shots].sort((a, b) => a.shot_number - b.shot_number))
  }, [shots])

  useEffect(() => {
    setShotsLoading(true)
    fetch(`/shots/hole/${hole.id}`)
      .then(r => r.json())
      .then((data: Shot[]) => { setShots(data); setShotsLoading(false) })
      .catch(() => setShotsLoading(false))
  }, [shotsBust, hole.id])

  async function saveStats() {
    setHoleBust(b => b + 1)
    setEditingStats(false)
    onSaved()
  }

  async function handleDrop(toId: number) {
    const fromId = draggedId.current
    setDragOverId(null)
    draggedId.current = null
    if (fromId === null || fromId === toId) return

    const fromIdx = localShots.findIndex(s => s.id === fromId)
    const toIdx   = localShots.findIndex(s => s.id === toId)
    if (fromIdx === -1 || toIdx === -1) return

    const next = [...localShots]
    const [moved] = next.splice(fromIdx, 1)
    next.splice(toIdx, 0, moved)
    setLocalShots(next)

    try {
      const r = await fetch(`/shots/hole/${hole.id}/reorder`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids: next.map(s => s.id) }),
      })
      if (!r.ok) throw new Error(await r.text())
      setReorderErr(null)
      setShotsBust(b => b + 1)
    } catch (e: any) {
      setReorderErr(e.message)
      setShotsBust(b => b + 1)
    }
  }

  function refreshShots() { setShotsBust(b => b + 1); setAdding(false) }

  return (
    <div>
      <div className="hole-detail-nav">
        <button className="btn-back" onClick={onBack}>← Scorecard</button>
        <div className="hole-nav-arrows">
          <button className="btn-sm hole-nav-arrow" onClick={onPrev} disabled={!onPrev}>← Prev</button>
          <button className="btn-sm hole-nav-arrow" onClick={onNext} disabled={!onNext}>Next →</button>
        </div>
      </div>

      <div className="hole-detail-head">
        <div>
          <div className="hole-detail-eyebrow">Hole</div>
          <div className="hole-detail-num">{current.hole_number}</div>
        </div>
        <div className="hole-detail-score-wrap">
          <span className={`hole-detail-score ${scoreClass(current.points)}`}>
            {current.score ?? '—'}
          </span>
          <span className="hole-detail-score-label">Score</span>
        </div>
      </div>

      {editingStats
        ? <HoleStatsForm
            hole={current}
            onSaved={saveStats}
            onCancel={() => setEditingStats(false)}
          />
        : <HoleStatsStrip hole={current} onEdit={() => setEditingStats(true)} />
      }

      <div className="section-header">
        <h2>Shots</h2>
        {!adding && (
          <button className="btn-sm btn-primary" onClick={() => setAdding(true)}>+ Add shot</button>
        )}
      </div>

      {reorderErr && <p className="error" style={{ marginBottom: 12 }}>{reorderErr}</p>}
      {shotsLoading && <p className="status">Loading shots…</p>}

      {!shotsLoading && localShots.length === 0 && !adding && (
        <p className="placeholder">No shots recorded for this hole.</p>
      )}

      {localShots.map(s => (
        <ShotRow
          key={s.id}
          shot={s}
          clubs={clubs}
          isDragOver={dragOverId === s.id}
          onDragStart={() => { draggedId.current = s.id }}
          onDragOver={e => { e.preventDefault(); setDragOverId(s.id) }}
          onDragLeave={() => setDragOverId(null)}
          onDrop={() => handleDrop(s.id)}
          onDragEnd={() => { draggedId.current = null; setDragOverId(null) }}
          onSaved={refreshShots}
          onDeleted={refreshShots}
        />
      ))}

      {adding && (
        <AddShotRow
          holeId={hole.id}
          nextShotNum={localShots.length + 1}
          clubs={clubs}
          onSaved={refreshShots}
          onCancel={() => setAdding(false)}
        />
      )}
    </div>
  )
}

// ── Scorecard row (display-only, navigates to hole detail) ────────────────────

function ScorecardRow({ hole, par, onSelect }: { hole: Hole; par: number | null; onSelect: () => void }) {
  return (
    <tbody>
      <tr className="sc-hole-row" onClick={onSelect}>
        <td className="sc-hole-num">{hole.hole_number}</td>
        <td className="sc-par-cell">{par ?? '—'}</td>
        <td className={`sc-score-cell ${scoreClass(hole.points)}`}>{hole.score ?? '—'}</td>
        <td className="sc-stat-cell">{hole.points ?? '—'}</td>
        <td className="sc-stat-cell">{hole.putts  ?? '—'}</td>
        <td className="sc-bool-cell">{boolLabel(hole.gir)}</td>
        <td className="sc-bool-cell">{boolLabel(hole.scramble_save)}</td>
        <td className="sc-bool-cell">{boolLabel(hole.penalty)}</td>
        <td className="sc-chevron-cell">›</td>
      </tr>
    </tbody>
  )
}

// ── Scorecard ─────────────────────────────────────────────────────────────────

function Scorecard({ holes, parMap, onSelectHole }: {
  holes: Hole[]
  parMap: ParMap
  onSelectHole: (h: Hole) => void
}) {
  const sorted = [...holes].sort((a, b) => a.hole_number - b.hole_number)
  const front  = sorted.filter(h => h.hole_number <= 9)
  const back   = sorted.filter(h => h.hole_number > 9)

  function parTotal(hs: Hole[]) {
    const vals = hs.map(h => parMap[h.course_hole_id]).filter((v): v is number => v != null)
    return vals.length > 0 ? vals.reduce((a, b) => a + b, 0) : '—'
  }

  return (
    <div className="scorecard-wrap">
      <table className="scorecard-table">
        <thead>
          <tr className="sc-head-row">
            <th className="sc-num-col">#</th>
            <th>Par</th>
            <th>Score</th>
            <th>Points</th>
            <th>Putts</th>
            <th>GIR</th>
            <th>Scramble</th>
            <th>Penalty</th>
            <th></th>
          </tr>
        </thead>

        {front.map(h => (
          <ScorecardRow key={h.id} hole={h}
            par={parMap[h.course_hole_id] ?? null}
            onSelect={() => onSelectHole(h)} />
        ))}

        {front.length > 0 && (
          <tbody>
            <tr className="sc-sub-row">
              <td className="sc-sub-label">OUT</td>
              <td className="sc-sub-val sc-par-sub">{parTotal(front)}</td>
              <td className="sc-sub-val">{sum(front, 'score')}</td>
              <td className="sc-sub-val">{sum(front, 'points')}</td>
              <td className="sc-sub-val">{sum(front, 'putts')}</td>
              <td colSpan={4}></td>
            </tr>
          </tbody>
        )}

        {back.map(h => (
          <ScorecardRow key={h.id} hole={h}
            par={parMap[h.course_hole_id] ?? null}
            onSelect={() => onSelectHole(h)} />
        ))}

        {back.length > 0 && (
          <tbody>
            <tr className="sc-sub-row">
              <td className="sc-sub-label">IN</td>
              <td className="sc-sub-val sc-par-sub">{parTotal(back)}</td>
              <td className="sc-sub-val">{sum(back, 'score')}</td>
              <td className="sc-sub-val">{sum(back, 'points')}</td>
              <td className="sc-sub-val">{sum(back, 'putts')}</td>
              <td colSpan={4}></td>
            </tr>
            <tr className="sc-total-row">
              <td className="sc-total-label">TOTAL</td>
              <td className="sc-total-val sc-par-total">{parTotal(sorted)}</td>
              <td className="sc-total-val">{sum(sorted, 'score')}</td>
              <td className="sc-total-val">{sum(sorted, 'points')}</td>
              <td className="sc-total-val">{sum(sorted, 'putts')}</td>
              <td colSpan={4}></td>
            </tr>
          </tbody>
        )}
      </table>
    </div>
  )
}

// ── Round detail ──────────────────────────────────────────────────────────────

function RoundDetail({ round: initialRound, playerMap, courseMap, onBack, onNavigate }: {
  round: Round
  playerMap: NameMap
  courseMap: NameMap
  onBack: () => void
  onNavigate: (t: NavTarget) => void
}) {
  const [holesBust, setHolesBust]     = useState(0)
  const [roundBust, setRoundBust]     = useState(0)
  const [selectedHole, setSelectedHole] = useState<Hole | null>(null)

  const { data: holes, loading, error } = useJson<Hole[]>(`/holes/round/${initialRound.id}`, holesBust)
  const { data: freshRound }            = useJson<Round>(`/rounds/${initialRound.id}`, roundBust)
  const { data: rawClubs }             = useJson<Club[]>(`/clubs/player/${initialRound.player_id}/active`)
  const { data: tee }                  = useJson<Tee>(`/tees/course/${initialRound.course_id}/${encodeURIComponent(initialRound.tees)}`)
  const { data: teeHoles }             = useJson<TeeHole[]>(tee ? `/tee-holes/tee/${tee.id}` : null)

  const round = freshRound ?? initialRound
  const clubs = rawClubs
    ? [...rawClubs].sort((a, b) => (b.carry_reliable ?? -Infinity) - (a.carry_reliable ?? -Infinity))
    : []
  const parMap: ParMap = teeHoles
    ? Object.fromEntries(teeHoles.map(th => [th.course_hole_id, th.par]))
    : {}

  // Keep selectedHole in sync when holes refresh
  useEffect(() => {
    if (holes && selectedHole) {
      const updated = holes.find(h => h.id === selectedHole.id)
      if (updated) setSelectedHole(updated)
    }
  }, [holes]) // eslint-disable-line react-hooks/exhaustive-deps

  async function onHoleSaved() {
    await fetch(`/rounds/${round.id}/totals`, { method: 'PATCH' })
    setHolesBust(b => b + 1)
    setRoundBust(b => b + 1)
  }

  if (selectedHole) {
    const sorted     = holes ? [...holes].sort((a, b) => a.hole_number - b.hole_number) : []
    const idx        = sorted.findIndex(h => h.id === selectedHole.id)
    const prevHole   = idx > 0 ? sorted[idx - 1] : undefined
    const nextHole   = idx >= 0 && idx < sorted.length - 1 ? sorted[idx + 1] : undefined
    return (
      <HoleDetail
        hole={selectedHole}
        clubs={clubs}
        onBack={() => setSelectedHole(null)}
        onSaved={onHoleSaved}
        onPrev={prevHole ? () => setSelectedHole(prevHole) : undefined}
        onNext={nextHole ? () => setSelectedHole(nextHole) : undefined}
      />
    )
  }

  return (
    <div>
      <button className="btn-back" onClick={onBack}>← Rounds</button>

      <div className="round-hero">
        <div className="round-hero-left">
          <div className="round-hero-eyebrow">{fmtDate(round.played_at)}</div>
          <div className="round-hero-venue">
            <a href="#" className="round-hero-venue-link"
              onClick={e => { e.preventDefault(); onNavigate({ section: 'Courses', entityId: round.course_id }) }}>
              {courseMap[round.course_id] ?? `Course ${round.course_id}`}
            </a>
          </div>
          <div className="round-hero-meta">
            <a href="#" className="round-hero-player-link"
              onClick={e => { e.preventDefault(); onNavigate({ section: 'Players', entityId: round.player_id }) }}>
              {playerMap[round.player_id] ?? `Player ${round.player_id}`}
            </a>
            <span className="round-hero-sep">·</span>
            <span>{round.tees}</span>
            <span className="round-hero-sep">·</span>
            <span style={{ textTransform: 'capitalize' }}>{round.round_type}</span>
          </div>
        </div>
        <div className="round-hero-right">
          <div className="round-hero-score-main">
            <span className="round-hero-score-value">{round.total_score ?? '—'}</span>
            <span className="round-hero-score-label">Gross</span>
          </div>
          <div className="round-hero-score-stack">
            <div className="round-hero-score-minor">
              <span className="round-hero-score-minor-value">{round.total_points ?? '—'}</span>
              <span className="round-hero-score-minor-label">Points</span>
            </div>
            <div className="round-hero-score-minor">
              <span className="round-hero-score-minor-value">{round.total_putts ?? '—'}</span>
              <span className="round-hero-score-minor-label">Putts</span>
            </div>
          </div>
        </div>
      </div>

      <Status loading={loading} error={error} />

      {holes && holes.length === 0 && (
        <p className="placeholder">No holes recorded for this round.</p>
      )}
      {holes && holes.length > 0 && (
        <Scorecard holes={holes} parMap={parMap} onSelectHole={setSelectedHole} />
      )}
    </div>
  )
}

// ── Rounds dashboard ──────────────────────────────────────────────────────────

function RoundsDashboard({ playerMap, courseMap, onSelectRound, onNavigate }: {
  playerMap: NameMap
  courseMap: NameMap
  onSelectRound: (r: Round) => void
  onNavigate: (t: NavTarget) => void
}) {
  const { data, loading, error } = useJson<Round[]>('/rounds/')
  const sorted = data ? [...data].sort((a, b) => b.played_at.localeCompare(a.played_at)) : null

  return (
    <>
      <Status loading={loading} error={error} />
      {sorted && sorted.length === 0 && <p className="placeholder">No rounds recorded.</p>}
      {sorted && sorted.length > 0 && (
        <div className="rounds-grid">
          {sorted.map(r => (
            <div key={r.id} className="round-card" onClick={() => onSelectRound(r)}>
              <span className="round-card-date">{fmtDate(r.played_at)}</span>
              <div className="round-card-who">
                <a href="#" className="round-card-link"
                  onClick={e => { e.stopPropagation(); e.preventDefault(); onNavigate({ section: 'Players', entityId: r.player_id }) }}>
                  {playerMap[r.player_id] ?? `Player ${r.player_id}`}
                </a>
                <span className="round-card-sep">·</span>
                <a href="#" className="round-card-link"
                  onClick={e => { e.stopPropagation(); e.preventDefault(); onNavigate({ section: 'Courses', entityId: r.course_id }) }}>
                  {courseMap[r.course_id] ?? `Course ${r.course_id}`}
                </a>
              </div>
              <div className="round-card-tags">
                <span className="round-tag">{r.tees}</span>
                <span className="round-tag">{r.round_type}</span>
              </div>
              <div className="round-card-stats">
                <div className="round-card-stat">
                  <span className="round-card-stat-value">{r.total_score  ?? '—'}</span>
                  <span className="round-card-stat-label">Score</span>
                </div>
                <div className="round-card-stat">
                  <span className="round-card-stat-value">{r.total_points ?? '—'}</span>
                  <span className="round-card-stat-label">Points</span>
                </div>
                <div className="round-card-stat">
                  <span className="round-card-stat-value">{r.total_putts  ?? '—'}</span>
                  <span className="round-card-stat-label">Putts</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  )
}

// ── Section root ──────────────────────────────────────────────────────────────

export function RoundsSection({ onNavigate }: { onNavigate: (t: NavTarget) => void }) {
  const [selectedRound, setSelectedRound] = useState<Round | null>(null)
  const { data: players } = useJson<Player[]>('/players/')
  const { data: courses } = useJson<Course[]>('/courses/')

  const playerMap: NameMap = players ? Object.fromEntries(players.map(p => [p.id, p.name])) : {}
  const courseMap: NameMap = courses ? Object.fromEntries(courses.map(c => [c.id, c.name])) : {}

  if (selectedRound) {
    return (
      <RoundDetail
        round={selectedRound}
        playerMap={playerMap}
        courseMap={courseMap}
        onBack={() => setSelectedRound(null)}
        onNavigate={t => { setSelectedRound(null); onNavigate(t) }}
      />
    )
  }

  return (
    <>
      <h1>Rounds</h1>
      <RoundsDashboard
        playerMap={playerMap}
        courseMap={courseMap}
        onSelectRound={setSelectedRound}
        onNavigate={onNavigate}
      />
    </>
  )
}
