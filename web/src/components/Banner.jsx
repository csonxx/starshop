import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import './Banner.css'

export default function Banner({ items }) {
  const [idx, setIdx] = useState(0)

  useEffect(() => {
    if (!items?.length) return
    setIdx((i) => i % items.length)
    const t = setInterval(() => setIdx((i) => (i + 1) % items.length), 5500)
    return () => clearInterval(t)
  }, [items])

  if (!items?.length) {
    return (
      <div className="banner banner-skeleton">
        <div className="skeleton" style={{ width: '100%', height: '100%' }} />
      </div>
    )
  }

  return (
    <section className="banner">
      {items.map((it, i) => (
        <div key={it.id || i} className={`bn-slide ${i === idx ? 'active' : ''}`}>
          <img src={it.image} alt={it.title} loading={i === 0 ? 'eager' : 'lazy'} />
          <div className="bn-mask" />
          <div className="bn-content container">
            <div className="bn-eyebrow display">ATELIER · FACTORY DIRECT</div>
            <h1 className="bn-title">{it.title}</h1>
            <p className="bn-sub">{it.subtitle}</p>
            <div className="bn-actions">
              <Link to={it.link || '/cases'} className="btn btn-gold">探索系列</Link>
              <Link to="/cases#cases" className="btn btn-ghost bn-ghost-light">查看案例</Link>
            </div>
          </div>
        </div>
      ))}
      <div className="bn-dots">
        {items.map((_, i) => (
          <button
            key={i}
            className={`bn-dot ${i === idx ? 'active' : ''}`}
            onClick={() => setIdx(i)}
            aria-label={`banner-${i + 1}`}
          />
        ))}
      </div>
    </section>
  )
}
