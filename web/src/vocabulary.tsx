import { useState } from 'react'
import { useJson, Status } from './hooks'

type VocabEntry = {
  domain: string
  value: string
  label: string
  sort_order: number
}

type DomainDraft = {
  value: string
  label: string
  sort_order: string
}

const DOMAIN_LABELS: Record<string, string> = {
  shot_type:        'Shot Type',
  shot_result:      'Shot Result',
  shot_miss:        'Miss Direction',
  shot_strike:      'Strike Quality',
  shot_source:      'Shot Source',
  round_type:       'Round Type',
  competition_type: 'Competition Type',
  flag_position:    'Flag Position',
  poi_type:         'POI Type',
  poi_side:         'POI Side',
}

// ── Entry row (view + edit) ───────────────────────────────────────────────────

function EntryRow({ entry, onSaved, onDeleted }: {
  entry: VocabEntry
  onSaved: () => void
  onDeleted: () => void
}) {
  const [editing, setEditing]       = useState(false)
  const [label, setLabel]           = useState(entry.label)
  const [sortOrder, setSortOrder]   = useState(String(entry.sort_order))
  const [saving, setSaving]         = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [deleting, setDeleting]     = useState(false)
  const [error, setError]           = useState<string | null>(null)

  async function save() {
    if (!label.trim()) { setError('Label is required'); return }
    setSaving(true); setError(null)
    try {
      const r = await fetch(`/vocabulary/${entry.domain}/${encodeURIComponent(entry.value)}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ label: label.trim(), sort_order: parseInt(sortOrder) || 0 }),
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
      const r = await fetch(`/vocabulary/${entry.domain}/${encodeURIComponent(entry.value)}`, {
        method: 'DELETE',
      })
      if (!r.ok) throw new Error(await r.text())
      onDeleted()
    } catch (e: any) {
      setError(e.message)
      setDeleting(false)
      setConfirming(false)
    }
  }

  if (!editing) {
    return (
      <tr>
        <td className="vocab-value-cell">{entry.value}</td>
        <td>{entry.label}</td>
        <td className="vocab-order-cell">{entry.sort_order}</td>
        <td>
          {!confirming ? (
            <>
              <button className="btn-sm" onClick={() => { setLabel(entry.label); setSortOrder(String(entry.sort_order)); setEditing(true) }}>Edit</button>
              {' '}
              <button className="btn-sm btn-danger" onClick={() => setConfirming(true)}>Delete</button>
            </>
          ) : (
            <>
              <span className="confirm-label">Delete?</span>
              <button className="btn-sm btn-danger" onClick={doDelete} disabled={deleting}>{deleting ? '…' : 'Yes'}</button>
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
      <td className="vocab-value-cell">{entry.value}</td>
      <td>
        <input className="cell-input" style={{ width: 180 }} value={label}
          onChange={e => setLabel(e.target.value)}
          onKeyDown={e => { if (e.key === 'Enter') save(); if (e.key === 'Escape') setEditing(false) }}
          autoFocus />
      </td>
      <td>
        <input className="cell-input" type="number" style={{ width: 64 }} value={sortOrder}
          onChange={e => setSortOrder(e.target.value)} />
      </td>
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        {' '}
        <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>Cancel</button>
        {error && <div className="inline-error" style={{ marginTop: 4 }}>{error}</div>}
      </td>
    </tr>
  )
}

// ── Add entry row ─────────────────────────────────────────────────────────────

function AddEntryRow({ domain, nextSortOrder, onSaved, onCancel }: {
  domain: string
  nextSortOrder: number
  onSaved: () => void
  onCancel: () => void
}) {
  const [draft, setDraft] = useState<DomainDraft>({ value: '', label: '', sort_order: String(nextSortOrder) })
  const [saving, setSaving] = useState(false)
  const [error, setError]   = useState<string | null>(null)

  async function save() {
    if (!draft.value.trim() || !draft.label.trim()) { setError('Value and label are required'); return }
    setSaving(true); setError(null)
    try {
      const r = await fetch('/vocabulary', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          domain,
          value:      draft.value.trim().toLowerCase().replace(/\s+/g, '_'),
          label:      draft.label.trim(),
          sort_order: parseInt(draft.sort_order) || nextSortOrder,
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
        <input className="cell-input" style={{ width: 130 }} placeholder="value"
          value={draft.value} onChange={e => setDraft(d => ({ ...d, value: e.target.value }))}
          autoFocus />
      </td>
      <td>
        <input className="cell-input" style={{ width: 180 }} placeholder="Display label"
          value={draft.label} onChange={e => setDraft(d => ({ ...d, label: e.target.value }))}
          onKeyDown={e => { if (e.key === 'Enter') save() }} />
      </td>
      <td>
        <input className="cell-input" type="number" style={{ width: 64 }}
          value={draft.sort_order} onChange={e => setDraft(d => ({ ...d, sort_order: e.target.value }))} />
      </td>
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Add'}</button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>Cancel</button>
        {error && <div className="inline-error" style={{ marginTop: 4 }}>{error}</div>}
      </td>
    </tr>
  )
}

// ── Domain section ────────────────────────────────────────────────────────────

function DomainSection({ domain, entries, onChanged }: {
  domain: string
  entries: VocabEntry[]
  onChanged: () => void
}) {
  const [adding, setAdding] = useState(false)
  const sorted = [...entries].sort((a, b) => a.sort_order - b.sort_order)
  const nextSortOrder = sorted.length > 0 ? sorted[sorted.length - 1].sort_order + 1 : 1

  function handleSaved() { setAdding(false); onChanged() }

  return (
    <div className="vocab-domain">
      <div className="section-header">
        <h2>{DOMAIN_LABELS[domain] ?? domain}</h2>
        {!adding && (
          <button className="btn-sm btn-primary" onClick={() => setAdding(true)}>+ Add</button>
        )}
      </div>
      <table>
        <thead>
          <tr>
            <th>Value</th>
            <th>Label</th>
            <th>Order</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {sorted.map(e => (
            <EntryRow key={e.value} entry={e} onSaved={onChanged} onDeleted={onChanged} />
          ))}
          {adding && (
            <AddEntryRow
              domain={domain}
              nextSortOrder={nextSortOrder}
              onSaved={handleSaved}
              onCancel={() => setAdding(false)}
            />
          )}
        </tbody>
      </table>
    </div>
  )
}

// ── Section root ──────────────────────────────────────────────────────────────

export function VocabularySection() {
  const [bust, setBust]           = useState(0)
  const [activeDomain, setActiveDomain] = useState<string | null>(null)
  const { data, loading, error }  = useJson<VocabEntry[]>('/vocabulary/', bust)

  const byDomain: Record<string, VocabEntry[]> = {}
  if (data) {
    for (const e of data) {
      ;(byDomain[e.domain] ??= []).push(e)
    }
  }

  const domains = Object.keys(byDomain).sort((a, b) => {
    const ai = Object.keys(DOMAIN_LABELS).indexOf(a)
    const bi = Object.keys(DOMAIN_LABELS).indexOf(b)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })

  const current = activeDomain ?? domains[0] ?? null

  return (
    <>
      <h1>Vocabulary</h1>
      <p className="vocab-subtitle">
        Controlled values stored in the database. The agent queries these before writing to any enumerated field.
      </p>

      <Status loading={loading} error={error} />

      {domains.length > 0 && (
        <>
          <div className="entity-tabs">
            {domains.map(d => (
              <button
                key={d}
                className={d === current ? 'active' : ''}
                onClick={() => setActiveDomain(d)}
              >
                {DOMAIN_LABELS[d] ?? d}
              </button>
            ))}
          </div>

          {current && byDomain[current] && (
            <DomainSection
              key={current}
              domain={current}
              entries={byDomain[current]}
              onChanged={() => setBust(b => b + 1)}
            />
          )}
        </>
      )}
    </>
  )
}
