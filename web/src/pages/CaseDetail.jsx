import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api } from '../api'
import { useUser } from '../store/user.jsx'
import './CaseDetail.css'

export default function CaseDetail() {
  const { id } = useParams()
  const nav = useNavigate()
  const { user } = useUser()
  const [item, setItem] = useState(null)
  const [loading, setLoading] = useState(true)
  const [imgIdx, setImgIdx] = useState(0)
  const [saved, setSaved] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    setLoading(true)
    setItem(null)
    setImgIdx(0)
    setError('')
    api.caseDetail(id, { signal: controller.signal })
      .then((r) => {
        setItem(r.data)
      })
      .catch((e) => {
        if (e.name !== 'CanceledError') setError(e.message || '案例加载失败')
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    window.scrollTo({ top: 0, behavior: 'smooth' })
    return () => controller.abort()
  }, [id])

  useEffect(() => {
    const savedCases = readSavedCases()
    setSaved(savedCases.includes(id))
  }, [id])

  if (loading) return <div className="cd-loading">载入中...</div>
  if (error) {
    return (
      <div className="cd-loading cd-error" role="alert">
        <span>{error}</span>
        <button className="btn btn-ghost" onClick={() => window.location.reload()}>重新加载</button>
      </div>
    )
  }
  if (!item) return <div className="cd-loading">案例不存在或已下架</div>

  const allImgs = [item.cover, ...(item.images || [])].filter(Boolean)

  const bookMeasure = () => {
    nav(user ? '/me?intent=measure' : `/login?redirect=${encodeURIComponent('/me?intent=measure')}`)
  }

  const toggleSaved = () => {
    if (!user) {
      nav(`/login?redirect=${encodeURIComponent(`/cases/${id}`)}`)
      return
    }
    const savedCases = readSavedCases()
    const next = savedCases.includes(id)
      ? savedCases.filter((caseID) => caseID !== id)
      : [...savedCases, id]
    localStorage.setItem('star_saved_cases', JSON.stringify(next))
    setSaved(next.includes(id))
  }

  return (
    <div className="cd">
      <div className="cd-back container">
        <button className="cd-back-btn" onClick={() => nav(-1)}>← 返回</button>
      </div>

      <div className="cd-hero container">
        <div className="cd-hero-l">
          <div className="cd-eyebrow display">{item.style} · {item.space} · {item.size}</div>
          <h1 className="cd-title">{item.title}</h1>
          <p className="cd-sub">设计师 1V1 定制 · 工厂直营 · 全程可视化交付</p>

          <div className="cd-price-card">
            <div className="cd-price-row">
              <div>
                <div className="cd-price-label">{item.price ? '精准报价' : '参考价格'}</div>
                {item.price && item.price > 0 ? (
                  <div className="cd-price">
                    <span className="num-sym">¥</span>
                    <span className="num">{(item.price / 10000).toFixed(2)}</span>
                    <em>万</em>
                  </div>
                ) : (
                  <div className="cd-price cd-price-locked">
                    <span className="num">{item.priceLabel || '请询价'}</span>
                    <small>登录销售/供应商可查看精准价</small>
                  </div>
                )}
                <div className="cd-price-range">{item.priceLabel} 区间</div>
              </div>
              <div className="cd-cta-group">
                <button className="btn btn-gold" onClick={bookMeasure}>立即预约量尺</button>
                <button className="btn btn-ghost" onClick={toggleSaved}>{saved ? '已保存' : '保存到我的'}</button>
              </div>
            </div>
          </div>

          <div className="cd-spec">
            <Spec k="项目面积" v={item.area} />
            <Spec k="空间类型" v={item.space} />
            <Spec k="风格定位" v={item.style} />
            <Spec k="产品规格" v={item.size} />
            <Spec k="主色系" v={(item.colors || []).join(' / ')} />
          </div>
        </div>

        <div className="cd-hero-r">
          <div className="cd-gallery">
            <img src={allImgs[imgIdx]} alt={item.title} />
            <div className="cd-thumbs">
              {allImgs.map((src, i) => (
                <button
                  key={i}
                  className={`cd-thumb ${i === imgIdx ? 'active' : ''}`}
                  onClick={() => setImgIdx(i)}
                >
                  <img src={src} alt="" />
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      <section className="cd-block container">
        <h2 className="serif">设计亮点</h2>
        <ul className="cd-highlights">
          {(item.highlights || []).map((h, i) => (
            <li key={`${h}-${i}`}><span className="cd-num">{(i + 1).toString().padStart(2, '0')}</span>{h}</li>
          ))}
        </ul>
      </section>

      <section className="cd-block container cd-two">
        <div>
          <h2 className="serif">主材配置</h2>
          <ul className="cd-list">
            {(item.materials || []).map((m, i) => <li key={`${m}-${i}`}>{m}</li>)}
          </ul>
        </div>
        <div>
          <h2 className="serif">五金配置</h2>
          <ul className="cd-list">
            {(item.hardware || []).map((m, i) => <li key={`${m}-${i}`}>{m}</li>)}
          </ul>
        </div>
      </section>

      <section className="cd-cta">
        <div className="container cd-cta-row">
          <div>
            <h2 className="serif">想要同款方案？</h2>
            <p>留下手机号，设计师 1V1 与您沟通 · 到店免费量尺</p>
          </div>
          <button className="btn btn-gold btn-large" onClick={bookMeasure}>立即预约量尺</button>
        </div>
      </section>
    </div>
  )
}

function readSavedCases() {
  try {
    const value = JSON.parse(localStorage.getItem('star_saved_cases') || '[]')
    return Array.isArray(value) ? value : []
  } catch {
    return []
  }
}

function Spec({ k, v }) {
  return (
    <div className="cd-spec-row">
      <span className="cd-spec-k">{k}</span>
      <span className="cd-spec-v">{v}</span>
    </div>
  )
}
