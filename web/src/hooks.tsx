import { useState, useEffect } from 'react'

// ── Vocabulary cache ──────────────────────────────────────────────────────────

type VocabEntry   = { value: string; label: string }
type VocabMap     = Record<string, VocabEntry[]>
type RawVocabEntry = { domain: string; value: string; label: string; sort_order: number }

let vocabCache: VocabMap | null = null
let vocabFetch: Promise<VocabMap> | null = null

export function useVocabulary(): { vocab: VocabMap; loading: boolean } {
  const [vocab, setVocab]     = useState<VocabMap>(() => vocabCache ?? {})
  const [loading, setLoading] = useState(!vocabCache)

  useEffect(() => {
    if (vocabCache) return
    if (!vocabFetch) {
      vocabFetch = fetch('/vocabulary/')
        .then(r => r.json() as Promise<RawVocabEntry[]>)
        .then(entries => {
          const map: VocabMap = {}
          for (const e of entries) (map[e.domain] ??= []).push({ value: e.value, label: e.label })
          return (vocabCache = map)
        })
    }
    vocabFetch.then(m => { setVocab(m); setLoading(false) })
  }, [])

  return { vocab, loading }
}

// ── JSON fetcher ──────────────────────────────────────────────────────────────

export function useJson<T>(url: string | null, bust = 0) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(url !== null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!url) { setData(null); setLoading(false); return }
    setLoading(true)
    setError(null)
    fetch(url)
      .then(r => { if (!r.ok) throw new Error(`${r.status} ${r.statusText}`); return r.json() })
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [url, bust])

  return { data, loading, error }
}

export function Status({ loading, error }: { loading: boolean; error: string | null }) {
  if (loading) return <p className="status">Loading…</p>
  if (error)   return <p className="error">{error}</p>
  return null
}
