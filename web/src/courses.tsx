import { useState, useEffect } from 'react'
import { useJson, useVocabulary, Status } from './hooks'

// ── Types ─────────────────────────────────────────────────────────────────────

type Course = { id: number; name: string; golf_api_id: string | null; created_at: string }
type Tee = { id: number; course_id: number; name: string; slope_rating: number | null; course_rating: number | null }
type Hole = { id: number; course_id: number; hole_number: number; green_centre_lat: number | null; green_centre_lng: number | null }
type TeeHole = {
  id: number; course_hole_id: number; tee_id: number
  par: number; stroke_index: number | null; distance: number
  tee_centre_lat: number | null; tee_centre_lng: number | null
}
type POI = {
  id: number; course_hole_id: number; specific_tee: string | null
  poi_type: string; side: string | null; reference_point: string | null
  distance_start: number | null; distance_end: number | null; label: string
}

// ── useOptionalJson — returns null on 404, error only for other failures ──────

function useOptionalJson<T>(url: string, bust = 0) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [notFound, setNotFound] = useState(false)

  useEffect(() => {
    setLoading(true); setError(null); setNotFound(false)
    fetch(url)
      .then(r => {
        if (r.status === 404) { setNotFound(true); setData(null); return null }
        if (!r.ok) throw new Error(`${r.status} ${r.statusText}`)
        return r.json()
      })
      .then(d => { if (d != null) setData(d) })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [url, bust])

  return { data, loading, error, notFound }
}

// ── Courses list ──────────────────────────────────────────────────────────────

function CoursesList({ onSelect }: { onSelect: (id: number) => void }) {
  const { data, loading, error } = useJson<Course[]>('/courses/')
  return (
    <>
      <Status loading={loading} error={error} />
      {data && (
        <table>
          <thead>
            <tr><th>Name</th><th>Golf API ID</th><th>Created</th></tr>
          </thead>
          <tbody>
            {data.map(c => (
              <tr key={c.id}>
                <td><a href="#" onClick={e => { e.preventDefault(); onSelect(c.id) }}>{c.name}</a></td>
                <td>{c.golf_api_id ?? '—'}</td>
                <td>{new Date(c.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  )
}

// ── Course details editor ─────────────────────────────────────────────────────

function CourseDetailsEditor({ course, onSaved }: { course: Course; onSaved: () => void }) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(course.name)
  const [golfApiId, setGolfApiId] = useState(course.golf_api_id ?? '')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!editing) { setName(course.name); setGolfApiId(course.golf_api_id ?? '') }
  }, [course, editing])

  async function save() {
    if (!name.trim()) { setError('Name is required'); return }
    setSaving(true)
    setError(null)
    try {
      const body: Record<string, unknown> = { name: name.trim() }
      if (golfApiId.trim()) body.golf_api_id = golfApiId.trim()
      const r = await fetch(`/courses/${course.id}`, {
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

  if (!editing) {
    return (
      <div className="detail-fields">
        <div className="inline-field">
          <span className="field-label">Name</span>
          <span className="field-value">{course.name}</span>
        </div>
        <div className="inline-field">
          <span className="field-label">Golf API ID</span>
          <span className="field-value">{course.golf_api_id ?? '—'}</span>
          <button className="btn-sm" onClick={() => { setEditing(true); setError(null) }}>Edit</button>
        </div>
      </div>
    )
  }

  return (
    <div className="detail-fields">
      <div className="inline-field">
        <span className="field-label">Name</span>
        <input className="inline-input" style={{ width: 240 }} value={name} onChange={e => setName(e.target.value)} autoFocus />
      </div>
      <div className="inline-field">
        <span className="field-label">Golf API ID</span>
        <input className="inline-input" style={{ width: 180 }} value={golfApiId} onChange={e => setGolfApiId(e.target.value)} placeholder="optional" />
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
        <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>Cancel</button>
        {error && <span className="inline-error">{error}</span>}
      </div>
    </div>
  )
}

// ── Tees section ──────────────────────────────────────────────────────────────

type TeeDraft = { slope_rating: string; course_rating: string }

function TeeRow({
  tee, editing, onEdit, onCancel, onSaved, onDeleted,
}: {
  tee: Tee; editing: boolean
  onEdit: () => void; onCancel: () => void; onSaved: () => void; onDeleted: () => void
}) {
  const [draft, setDraft] = useState<TeeDraft>({
    slope_rating:  tee.slope_rating  != null ? String(tee.slope_rating)  : '',
    course_rating: tee.course_rating != null ? String(tee.course_rating) : '',
  })
  const [saving, setSaving] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (editing) {
      setDraft({
        slope_rating:  tee.slope_rating  != null ? String(tee.slope_rating)  : '',
        course_rating: tee.course_rating != null ? String(tee.course_rating) : '',
      })
      setError(null)
    }
  }, [editing]) // eslint-disable-line react-hooks/exhaustive-deps

  async function save() {
    setSaving(true); setError(null)
    try {
      const body: Record<string, unknown> = {}
      if (draft.slope_rating !== '')  body.slope_rating  = parseInt(draft.slope_rating)
      if (draft.course_rating !== '') body.course_rating = parseFloat(draft.course_rating)
      const r = await fetch(`/tees/${tee.id}`, {
        method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  async function del() {
    setSaving(true)
    try {
      const r = await fetch(`/tees/${tee.id}`, { method: 'DELETE' })
      if (!r.ok) throw new Error(await r.text())
      onDeleted()
    } catch (e: any) { setError(e.message); setSaving(false); setConfirming(false) }
  }

  if (!editing) {
    return (
      <tr>
        <td>{tee.name}</td>
        <td>{tee.slope_rating ?? '—'}</td>
        <td>{tee.course_rating ?? '—'}</td>
        <td>
          {!confirming ? (
            <>
              <button className="btn-sm" onClick={onEdit}>Edit</button>
              {' '}
              <button className="btn-sm btn-danger" onClick={() => setConfirming(true)}>Delete</button>
            </>
          ) : (
            <>
              <span className="confirm-label">Delete tee?</span>
              <button className="btn-sm btn-danger" onClick={del} disabled={saving}>{saving ? '…' : 'Yes'}</button>
              {' '}
              <button className="btn-sm" onClick={() => setConfirming(false)}>No</button>
            </>
          )}
          {error && <div className="inline-error">{error}</div>}
        </td>
      </tr>
    )
  }

  return (
    <tr className="row-editing">
      <td><strong>{tee.name}</strong></td>
      <td><input className="cell-input" type="number" value={draft.slope_rating}  onChange={e => setDraft(d => ({ ...d, slope_rating: e.target.value }))}  placeholder="—" /></td>
      <td><input className="cell-input" type="number" step="0.1" value={draft.course_rating} onChange={e => setDraft(d => ({ ...d, course_rating: e.target.value }))} placeholder="—" /></td>
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error">{error}</div>}
      </td>
    </tr>
  )
}

function AddTeeRow({ courseId, onCancel, onSaved }: { courseId: number; onCancel: () => void; onSaved: () => void }) {
  const [name, setName] = useState('')
  const [slopeRating, setSlopeRating] = useState('')
  const [courseRating, setCourseRating] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function save() {
    if (!name.trim()) { setError('Name is required'); return }
    setSaving(true); setError(null)
    try {
      const body: Record<string, unknown> = { course_id: courseId, name: name.trim() }
      if (slopeRating !== '')  body.slope_rating  = parseInt(slopeRating)
      if (courseRating !== '') body.course_rating = parseFloat(courseRating)
      const r = await fetch('/tees', {
        method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  return (
    <tr className="row-adding">
      <td><input className="cell-input" type="text" value={name} onChange={e => setName(e.target.value)} placeholder="e.g. white" autoFocus style={{ width: 80 }} /></td>
      <td><input className="cell-input" type="number" value={slopeRating} onChange={e => setSlopeRating(e.target.value)} placeholder="—" /></td>
      <td><input className="cell-input" type="number" step="0.1" value={courseRating} onChange={e => setCourseRating(e.target.value)} placeholder="—" /></td>
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Add'}</button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error">{error}</div>}
      </td>
    </tr>
  )
}

function TeesSection({ courseId }: { courseId: number }) {
  const [bust, setBust] = useState(0)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)
  const { data: tees, loading, error } = useJson<Tee[]>(`/tees/course/${courseId}`, bust)

  function refresh() { setEditingId(null); setAdding(false); setBust(b => b + 1) }

  return (
    <>
      <div className="section-header">
        <h2>Tees</h2>
        {!adding && <button className="btn-sm btn-primary" onClick={() => { setEditingId(null); setAdding(true) }}>+ Add tee</button>}
      </div>
      <Status loading={loading} error={error} />
      {tees && (tees.length > 0 || adding) && (
        <table>
          <thead><tr><th>Name</th><th>Slope rating</th><th>Course rating</th><th></th></tr></thead>
          <tbody>
            {tees.map(t => (
              <TeeRow
                key={t.id} tee={t}
                editing={editingId === t.id}
                onEdit={() => { setAdding(false); setEditingId(t.id) }}
                onCancel={() => setEditingId(null)}
                onSaved={refresh} onDeleted={refresh}
              />
            ))}
            {adding && <AddTeeRow courseId={courseId} onCancel={() => setAdding(false)} onSaved={refresh} />}
          </tbody>
        </table>
      )}
      {tees && tees.length === 0 && !adding && <p className="placeholder">No tees recorded.</p>}
    </>
  )
}

// ── Holes — green centre editor ───────────────────────────────────────────────

function GreenCentreEditor({ hole, onSaved }: { hole: Hole; onSaved: () => void }) {
  const [editing, setEditing] = useState(false)
  const [lat, setLat] = useState('')
  const [lng, setLng] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    setLat(hole.green_centre_lat != null ? String(hole.green_centre_lat) : '')
    setLng(hole.green_centre_lng != null ? String(hole.green_centre_lng) : '')
    setEditing(true); setError(null)
  }

  async function save() {
    setSaving(true); setError(null)
    try {
      const body: Record<string, unknown> = {}
      if (lat.trim() !== '') body.green_centre_lat = parseFloat(lat)
      if (lng.trim() !== '') body.green_centre_lng = parseFloat(lng)
      const r = await fetch(`/course-holes/${hole.id}/coordinates`, {
        method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body),
      })
      if (!r.ok) throw new Error(await r.text())
      setEditing(false); onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  if (!editing) {
    const hasCoords = hole.green_centre_lat != null && hole.green_centre_lng != null
    const coords = hasCoords
      ? `${hole.green_centre_lat!.toFixed(6)}, ${hole.green_centre_lng!.toFixed(6)}`
      : '—'
    const mapsUrl = hasCoords
      ? `https://www.google.com/maps?q=${hole.green_centre_lat},${hole.green_centre_lng}`
      : null
    return (
      <div className="inline-field" style={{ marginBottom: 20 }}>
        <span className="field-label">Green centre</span>
        <span className="field-value" style={{ fontFamily: 'monospace', fontSize: 13 }}>{coords}</span>
        {mapsUrl && (
          <a href={mapsUrl} target="_blank" rel="noopener noreferrer" className="btn-link" style={{ fontSize: 12 }}>
            View on map ↗
          </a>
        )}
        <button className="btn-sm" onClick={startEdit}>Edit</button>
      </div>
    )
  }

  return (
    <div className="inline-field" style={{ marginBottom: 20, flexWrap: 'wrap', gap: 8 }}>
      <span className="field-label">Green centre</span>
      <input className="inline-input" style={{ width: 130 }} type="number" step="any" value={lat} onChange={e => setLat(e.target.value)} placeholder="lat" />
      <input className="inline-input" style={{ width: 130 }} type="number" step="any" value={lng} onChange={e => setLng(e.target.value)} placeholder="lng" />
      <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
      <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>Cancel</button>
      {error && <span className="inline-error">{error}</span>}
    </div>
  )
}

// ── Holes — POI table ─────────────────────────────────────────────────────────

type POIDraft = {
  poi_type: string; label: string; side: string
  reference_point: string; distance_start: string; distance_end: string; specific_tee: string
}

function poiToDraft(p: POI): POIDraft {
  return {
    poi_type:       p.poi_type,
    label:          p.label,
    side:           p.side           ?? '',
    reference_point: p.reference_point ?? '',
    distance_start: p.distance_start != null ? String(p.distance_start) : '',
    distance_end:   p.distance_end   != null ? String(p.distance_end)   : '',
    specific_tee:   p.specific_tee   ?? '',
  }
}

function emptyPOIDraft(): POIDraft {
  return { poi_type: '', label: '', side: '', reference_point: '', distance_start: '', distance_end: '', specific_tee: '' }
}

function draftToPOIBody(d: POIDraft): Record<string, unknown> {
  const body: Record<string, unknown> = { poi_type: d.poi_type, label: d.label }
  if (d.side !== '')             body.side            = d.side
  if (d.reference_point !== '')  body.reference_point = d.reference_point
  if (d.distance_start !== '')   body.distance_start  = parseFloat(d.distance_start)
  if (d.distance_end !== '')     body.distance_end    = parseFloat(d.distance_end)
  if (d.specific_tee !== '')     body.specific_tee    = d.specific_tee
  return body
}

function POIInputs({ draft, onChange }: {
  draft: POIDraft
  onChange: (key: keyof POIDraft) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => void
}) {
  const { vocab }   = useVocabulary()
  const poiTypes    = vocab['poi_type']        ?? []
  const poiSides    = vocab['poi_side']        ?? []
  const refPoints   = vocab['reference_point'] ?? []

  return (
    <>
      <td>
        <select className="cell-input" style={{ width: 110 }} value={draft.poi_type} onChange={onChange('poi_type')}>
          <option value="">—</option>
          {poiTypes.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </td>
      <td><input className="cell-input" style={{ width: 140 }} type="text" value={draft.label} onChange={onChange('label')} placeholder="label" /></td>
      <td>
        <select className="cell-input" style={{ width: 80 }} value={draft.side} onChange={onChange('side')}>
          <option value="">—</option>
          {poiSides.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </td>
      <td>
        <select className="cell-input" style={{ width: 80 }} value={draft.reference_point} onChange={onChange('reference_point')}>
          <option value="">—</option>
          {refPoints.map(e => <option key={e.value} value={e.value}>{e.label}</option>)}
        </select>
      </td>
      <td><input className="cell-input" style={{ width: 60 }} type="number" value={draft.distance_start} onChange={onChange('distance_start')} placeholder="—" /></td>
      <td><input className="cell-input" style={{ width: 60 }} type="number" value={draft.distance_end}   onChange={onChange('distance_end')}   placeholder="—" /></td>
      <td><input className="cell-input" style={{ width: 60 }} type="text"   value={draft.specific_tee}   onChange={onChange('specific_tee')}   placeholder="all" /></td>
    </>
  )
}

function POIRow({
  poi, editing, onEdit, onCancel, onSaved, onDeleted,
}: {
  poi: POI; editing: boolean
  onEdit: () => void; onCancel: () => void; onSaved: () => void; onDeleted: () => void
}) {
  const [draft, setDraft] = useState<POIDraft>(poiToDraft(poi))
  const [saving, setSaving] = useState(false)
  const [confirming, setConfirming] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (editing) { setDraft(poiToDraft(poi)); setError(null); setConfirming(false) }
  }, [editing]) // eslint-disable-line react-hooks/exhaustive-deps

  function field(key: keyof POIDraft) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    if (!draft.poi_type || !draft.label) { setError('Type and label required'); return }
    setSaving(true); setError(null)
    try {
      const r = await fetch(`/pois/${poi.id}`, {
        method: 'PUT', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(draftToPOIBody(draft)),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  async function del() {
    setSaving(true)
    try {
      const r = await fetch(`/pois/${poi.id}`, { method: 'DELETE' })
      if (!r.ok) throw new Error(await r.text())
      onDeleted()
    } catch (e: any) { setError(e.message); setSaving(false); setConfirming(false) }
  }

  if (!editing) {
    return (
      <tr>
        <td>{poi.poi_type}</td>
        <td>{poi.label}</td>
        <td>{poi.side ?? '—'}</td>
        <td>{poi.reference_point ?? '—'}</td>
        <td>{poi.distance_start ?? '—'}</td>
        <td>{poi.distance_end ?? '—'}</td>
        <td>{poi.specific_tee ?? 'all'}</td>
        <td>
          {!confirming ? (
            <>
              <button className="btn-sm" onClick={onEdit}>Edit</button>
              {' '}
              <button className="btn-sm btn-danger" onClick={() => setConfirming(true)}>✕</button>
            </>
          ) : (
            <>
              <span className="confirm-label">Delete?</span>
              <button className="btn-sm btn-danger" onClick={del} disabled={saving}>{saving ? '…' : 'Yes'}</button>
              {' '}
              <button className="btn-sm" onClick={() => setConfirming(false)}>No</button>
            </>
          )}
          {error && <div className="inline-error">{error}</div>}
        </td>
      </tr>
    )
  }

  return (
    <tr className="row-editing">
      <POIInputs draft={draft} onChange={field} />
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error">{error}</div>}
      </td>
    </tr>
  )
}

function AddPOIRow({ holeId, onCancel, onSaved }: { holeId: number; onCancel: () => void; onSaved: () => void }) {
  const [draft, setDraft] = useState<POIDraft>(emptyPOIDraft())
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function field(key: keyof POIDraft) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    if (!draft.poi_type || !draft.label) { setError('Type and label required'); return }
    setSaving(true); setError(null)
    try {
      const r = await fetch('/pois', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ course_hole_id: holeId, ...draftToPOIBody(draft) }),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  return (
    <tr className="row-adding">
      <POIInputs draft={draft} onChange={field} />
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Add'}</button>
        {' '}
        <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
        {error && <div className="inline-error">{error}</div>}
      </td>
    </tr>
  )
}

// ── Tee positions (per-tee row for a hole) ────────────────────────────────────

type TeeHoleDraft = { par: string; distance: string; stroke_index: string; lat: string; lng: string }

function teeHoleToDraft(th: TeeHole): TeeHoleDraft {
  return {
    par:          String(th.par),
    distance:     String(th.distance),
    stroke_index: th.stroke_index != null ? String(th.stroke_index) : '',
    lat:          th.tee_centre_lat != null ? String(th.tee_centre_lat) : '',
    lng:          th.tee_centre_lng != null ? String(th.tee_centre_lng) : '',
  }
}

function emptyTeeHoleDraft(): TeeHoleDraft {
  return { par: '', distance: '', stroke_index: '', lat: '', lng: '' }
}

function draftToTeeHoleBody(d: TeeHoleDraft): Record<string, unknown> {
  const body: Record<string, unknown> = {
    par:      parseInt(d.par)      || 4,
    distance: parseInt(d.distance) || 0,
  }
  if (d.stroke_index !== '') body.stroke_index   = parseInt(d.stroke_index)
  if (d.lat !== '')          body.tee_centre_lat = parseFloat(d.lat)
  if (d.lng !== '')          body.tee_centre_lng = parseFloat(d.lng)
  return body
}

function coordsLink(lat: number, lng: number) {
  return `https://www.google.com/maps?q=${lat},${lng}`
}

function TeeHoleRow({ tee, holeId }: { tee: Tee; holeId: number }) {
  const [bust, setBust] = useState(0)
  const { data: th, loading, error: fetchErr, notFound } = useOptionalJson<TeeHole>(
    `/tee-holes/hole/${holeId}/tee/${tee.id}`, bust
  )
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState<TeeHoleDraft>(emptyTeeHoleDraft())
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function startEdit() {
    setDraft(th ? teeHoleToDraft(th) : emptyTeeHoleDraft())
    setEditing(true); setError(null)
  }

  function field(key: keyof TeeHoleDraft) {
    return (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setDraft(d => ({ ...d, [key]: e.target.value }))
  }

  async function save() {
    if (!draft.par || !draft.distance) { setError('Par and distance are required'); return }
    setSaving(true); setError(null)
    try {
      const body = draftToTeeHoleBody(draft)
      const r = th
        ? await fetch(`/tee-holes/${th.id}`, { method: 'PUT',  headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) })
        : await fetch('/tee-holes',           { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ course_hole_id: holeId, tee_id: tee.id, ...body }) })
      if (!r.ok) throw new Error(await r.text())
      setEditing(false); setBust(b => b + 1)
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  if (loading) return <tr><td>{tee.name}</td><td colSpan={6} style={{ color: '#aaa', fontSize: 12 }}>Loading…</td></tr>
  if (fetchErr) return <tr><td>{tee.name}</td><td colSpan={6} className="inline-error">{fetchErr}</td></tr>

  if (!editing) {
    const hasCoords = th?.tee_centre_lat != null && th?.tee_centre_lng != null
    const coords = hasCoords
      ? `${th!.tee_centre_lat!.toFixed(6)}, ${th!.tee_centre_lng!.toFixed(6)}`
      : '—'
    return (
      <tr>
        <td>{tee.name}</td>
        <td>{th?.par ?? '—'}</td>
        <td>{th?.distance ?? '—'}</td>
        <td>{th?.stroke_index ?? '—'}</td>
        <td style={{ fontFamily: 'monospace', fontSize: 11 }}>{coords}</td>
        <td>
          {hasCoords && (
            <a href={coordsLink(th!.tee_centre_lat!, th!.tee_centre_lng!)}
               target="_blank" rel="noopener noreferrer" className="btn-link" style={{ fontSize: 11 }}>
              ↗
            </a>
          )}
        </td>
        <td>
          <button className="btn-sm" onClick={startEdit}>{notFound ? 'Add' : 'Edit'}</button>
          {error && <div className="inline-error">{error}</div>}
        </td>
      </tr>
    )
  }

  return (
    <tr className="row-editing">
      <td><strong>{tee.name}</strong></td>
      <td><input className="cell-input" style={{ width: 44 }} type="number" value={draft.par}          onChange={field('par')}          placeholder="4" /></td>
      <td><input className="cell-input" style={{ width: 56 }} type="number" value={draft.distance}     onChange={field('distance')}     placeholder="0" /></td>
      <td><input className="cell-input" style={{ width: 44 }} type="number" value={draft.stroke_index} onChange={field('stroke_index')} placeholder="—" /></td>
      <td><input className="cell-input" style={{ width: 110 }} type="number" step="any" value={draft.lat} onChange={field('lat')} placeholder="lat" /></td>
      <td><input className="cell-input" style={{ width: 110 }} type="number" step="any" value={draft.lng} onChange={field('lng')} placeholder="lng" /></td>
      <td>
        <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Save'}</button>
        {' '}
        <button className="btn-sm" onClick={() => setEditing(false)} disabled={saving}>✕</button>
        {error && <div className="inline-error">{error}</div>}
      </td>
    </tr>
  )
}

function TeePositionsSection({ holeId, tees }: { holeId: number; tees: Tee[] }) {
  if (tees.length === 0) return <p className="placeholder" style={{ padding: '12px 0' }}>No tees configured for this course.</p>
  return (
    <>
      <h4 style={{ margin: '20px 0 10px' }}>Tee positions</h4>
      <table className="poi-table">
        <thead>
          <tr><th>Tee</th><th>Par</th><th>Dist. (m)</th><th>SI</th><th>Tee centre</th><th></th><th></th></tr>
        </thead>
        <tbody>
          {tees.map(t => <TeeHoleRow key={t.id} tee={t} holeId={holeId} />)}
        </tbody>
      </table>
    </>
  )
}

// ── Hole detail (right panel) ─────────────────────────────────────────────────

function HoleDetail({ hole, tees, onHoleUpdated }: { hole: Hole; tees: Tee[]; onHoleUpdated: () => void }) {
  const [poiBust, setPoiBust] = useState(0)
  const [editingPoiId, setEditingPoiId] = useState<number | null>(null)
  const [addingPoi, setAddingPoi] = useState(false)
  const { data: pois, loading, error } = useJson<POI[]>(`/pois/hole/${hole.id}`, poiBust)

  function refreshPois() { setEditingPoiId(null); setAddingPoi(false); setPoiBust(b => b + 1) }

  return (
    <div className="hole-detail">
      <h3>Hole {hole.hole_number}</h3>

      <GreenCentreEditor hole={hole} onSaved={onHoleUpdated} />

      <TeePositionsSection holeId={hole.id} tees={tees} />

      <div className="section-header" style={{ marginTop: 20 }}>
        <h4>Points of interest</h4>
        {!addingPoi && (
          <button className="btn-sm btn-primary" onClick={() => { setEditingPoiId(null); setAddingPoi(true) }}>
            + Add POI
          </button>
        )}
      </div>

      <Status loading={loading} error={error} />
      {pois && (pois.length > 0 || addingPoi) && (
        <table className="poi-table">
          <thead>
            <tr>
              <th>Type</th><th>Label</th><th>Side</th><th>Ref.</th>
              <th>From (m)</th><th>To (m)</th><th>Tee</th><th></th>
            </tr>
          </thead>
          <tbody>
            {pois.map(p => (
              <POIRow
                key={p.id} poi={p}
                editing={editingPoiId === p.id}
                onEdit={() => { setAddingPoi(false); setEditingPoiId(p.id) }}
                onCancel={() => setEditingPoiId(null)}
                onSaved={refreshPois} onDeleted={refreshPois}
              />
            ))}
            {addingPoi && <AddPOIRow holeId={hole.id} onCancel={() => setAddingPoi(false)} onSaved={refreshPois} />}
          </tbody>
        </table>
      )}
      {pois && pois.length === 0 && !addingPoi && (
        <p className="placeholder" style={{ padding: '16px 0' }}>No POIs recorded for this hole.</p>
      )}
    </div>
  )
}

// ── Holes section (two-column master-detail) ──────────────────────────────────

function AddHoleForm({ courseId, nextNumber, onCancel, onSaved }: {
  courseId: number; nextNumber: number; onCancel: () => void; onSaved: () => void
}) {
  const [holeNum, setHoleNum] = useState(String(nextNumber))
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function save() {
    const n = parseInt(holeNum)
    if (isNaN(n) || n < 1) { setError('Valid hole number required'); return }
    setSaving(true); setError(null)
    try {
      const r = await fetch('/course-holes', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ course_id: courseId, hole_number: n }),
      })
      if (!r.ok) throw new Error(await r.text())
      onSaved()
    } catch (e: any) { setError(e.message) } finally { setSaving(false) }
  }

  return (
    <div className="add-hole-form">
      <input
        className="cell-input" type="number" min="1" max="27"
        value={holeNum} onChange={e => setHoleNum(e.target.value)}
        style={{ width: 56 }} autoFocus
      />
      <button className="btn-sm btn-primary" onClick={save} disabled={saving}>{saving ? '…' : 'Add'}</button>
      <button className="btn-sm" onClick={onCancel} disabled={saving}>✕</button>
      {error && <div className="inline-error">{error}</div>}
    </div>
  )
}

function HolesSection({ courseId }: { courseId: number }) {
  const [bust, setBust] = useState(0)
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [addingHole, setAddingHole] = useState(false)
  const { data: holes, loading, error } = useJson<Hole[]>(`/course-holes/course/${courseId}`, bust)
  const { data: tees } = useJson<Tee[]>(`/tees/course/${courseId}`)

  const sorted = holes ? [...holes].sort((a, b) => a.hole_number - b.hole_number) : null
  const selectedHole = sorted?.find(h => h.id === selectedId) ?? null
  const nextNumber = sorted ? (Math.max(0, ...sorted.map(h => h.hole_number)) + 1) : 1

  function refreshHoles() {
    setAddingHole(false)
    setBust(b => b + 1)
  }

  return (
    <>
      <div className="section-header">
        <h2>Holes</h2>
        {!addingHole && sorted && (
          <button className="btn-sm btn-primary" onClick={() => setAddingHole(true)}>+ Add hole</button>
        )}
      </div>
      <Status loading={loading} error={error} />
      {sorted && (
        <div className="holes-layout">
          <div className="holes-nav">
            {sorted.map(h => (
              <button
                key={h.id}
                className={`hole-nav-btn${selectedId === h.id ? ' active' : ''}`}
                onClick={() => setSelectedId(h.id)}
              >
                <span className="hole-num">{h.hole_number}</span>
                {h.green_centre_lat != null && <span className="hole-dot" title="Coordinates set" />}
              </button>
            ))}
            {addingHole && (
              <AddHoleForm
                courseId={courseId}
                nextNumber={nextNumber}
                onCancel={() => setAddingHole(false)}
                onSaved={() => { refreshHoles() }}
              />
            )}
            {sorted.length === 0 && !addingHole && (
              <p style={{ padding: '12px 16px', fontSize: 12, color: '#999' }}>No holes yet</p>
            )}
          </div>

          <div className="holes-detail-pane">
            {selectedHole ? (
              <HoleDetail
                key={selectedHole.id}
                hole={selectedHole}
                tees={tees ?? []}
                onHoleUpdated={() => setBust(b => b + 1)}
              />
            ) : (
              <p className="holes-empty-prompt">Select a hole to view and edit its details.</p>
            )}
          </div>
        </div>
      )}
    </>
  )
}

// ── Course profile ────────────────────────────────────────────────────────────

function CourseProfile({ id, onBack }: { id: number; onBack: () => void }) {
  const [bust, setBust] = useState(0)
  const { data: course, loading, error } = useJson<Course>(`/courses/${id}`, bust)

  return (
    <div>
      <button className="btn-back" onClick={onBack}>← Courses</button>

      {loading && <p className="status">Loading…</p>}
      {error   && <p className="error">{error}</p>}
      {course && (
        <>
          <h1 style={{ marginTop: 20, marginBottom: 16 }}>{course.name}</h1>
          <CourseDetailsEditor course={course} onSaved={() => setBust(b => b + 1)} />
          <TeesSection courseId={id} />
          <HolesSection courseId={id} />
        </>
      )}
    </div>
  )
}

// ── Section root ──────────────────────────────────────────────────────────────

export function CoursesSection({ initialId }: { initialId?: number }) {
  const [selectedId, setSelectedId] = useState<number | null>(initialId ?? null)

  if (selectedId !== null) {
    return <CourseProfile id={selectedId} onBack={() => setSelectedId(null)} />
  }

  return (
    <>
      <h1>Courses</h1>
      <CoursesList onSelect={setSelectedId} />
    </>
  )
}
