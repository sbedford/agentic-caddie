import { useState } from 'react'
import { PlayersSection } from './players'
import { CoursesSection } from './courses'
import { RoundsSection } from './rounds'
import { VocabularySection } from './vocabulary'
import type { Section, NavTarget } from './types'

const SECTIONS: Section[] = ['Players', 'Courses', 'Rounds', 'Vocabulary']

export default function App() {
  // nav.entityId is passed as initialId to the section so it opens that entity directly.
  // key={`${nav.section}-${nav.entityId}`} remounts the section component on cross-navigation,
  // ensuring useState(initialId) picks up the new value.
  const [nav, setNav] = useState<NavTarget>({ section: 'Players' })

  function navigate(target: NavTarget) {
    setNav(target)
  }

  function renderSection() {
    const key = `${nav.section}-${nav.entityId ?? 'none'}`
    switch (nav.section) {
      case 'Players':
        return <PlayersSection key={key} initialId={nav.entityId} />
      case 'Courses':
        return <CoursesSection key={key} initialId={nav.entityId} />
      case 'Rounds':
        return <RoundsSection key={key} onNavigate={navigate} />
      case 'Vocabulary':
        return <VocabularySection key={key} />
    }
  }

  return (
    <>
      <nav>
        <span className="logo">
          <svg className="logo-mark" viewBox="0 0 16 20" fill="none" aria-hidden="true">
            <line x1="3.5" y1="2" x2="3.5" y2="19" stroke="white" strokeWidth="1.5" strokeLinecap="round"/>
            <path d="M3.5 2.5 L14 7.5 L3.5 13 Z" fill="#4ade80"/>
            <circle cx="3.5" cy="19" r="1.5" fill="#4ade80"/>
          </svg>
          Agentic <span>Caddie</span>
        </span>
        {SECTIONS.map(s => (
          <button
            key={s}
            className={nav.section === s ? 'active' : ''}
            onClick={() => setNav({ section: s })}
          >
            {s}
          </button>
        ))}
      </nav>
      <main>
        {renderSection()}
      </main>
    </>
  )
}
