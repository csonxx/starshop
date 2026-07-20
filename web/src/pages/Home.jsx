import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import Banner from '../components/Banner'
import CaseCard from '../components/CaseCard'
import './Home.css'

export default function Home() {
  const [banners, setBanners] = useState([])
  const [styleTags, setStyleTags] = useState([])
  const [pinned, setPinned] = useState([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    Promise.all([
      api.banners(),
      api.tags('style'),
      api.pinned()
    ]).then(([b, st, pn]) => {
      setBanners(b.data || [])
      setStyleTags(st.data || [])
      setPinned(pn.data || [])
      setLoading(false)
    }).catch((e) => {
      console.error(e)
      setLoading(false)
    })
  }, [])

  return (
    <div className="home">
      <Banner items={banners} />

      <section className="factory" id="factory">
        <div className="container factory-row">
          <div className="factory-meta">
            <div className="eyebrow display">FACTORY DIRECT · 自有智造</div>
            <h2 className="serif">
              从<span className="hl"> 12,000㎡ </span>工厂<br />
              直送您的家
            </h2>
            <p>板材自产 · 五金进口 · 设计 1V1 · 28 天交付</p>
            <Link to="/cases" className="btn btn-gold" style={{ marginTop: 24 }}>
              进入案例库 →
            </Link>
          </div>
          <div className="factory-stats">
            <Stat n="28" unit="天" label="平均交付" />
            <Stat n="12,000" unit="㎡" label="自有厂房" />
            <Stat n="1V1" label="设计师全程跟进" />
            <Stat n="10" unit="年" label="质保" />
          </div>
        </div>
      </section>

      <section className="styles" id="styles">
        <div className="container">
          <SectionTitle
            eyebrow="01 · 风格 STYLE"
            title="甄选十一种主流风格"
            sub="点击任一风格 · 进入二级筛选页面"
          />
          <div className="style-tags">
            {styleTags.map((t) => (
              <Link
                key={t.id}
                to={`/style/${t.value}`}
                className="style-tag"
              >
                <span className="style-ic">{t.icon}</span>
                <span className="style-name">{t.name}</span>
              </Link>
            ))}
          </div>
        </div>
      </section>

      <section className="pinned-section">
        <div className="container">
          <SectionTitle
            eyebrow="02 · 置顶 HOT"
            title="运营精选案例"
            sub="后台可置顶 · 最多 8 个"
          />
          <div className="pinned-grid">
            {pinned.length === 0 && !loading && (
              <div className="empty">暂无置顶案例</div>
            )}
            {pinned.slice(0, 8).map((p) => <CaseCard key={p.id} item={p} />)}
          </div>
        </div>
      </section>

      <Footer />
    </div>
  )
}

function SectionTitle({ eyebrow, title, sub }) {
  return (
    <div className="section-title">
      <div className="eyebrow display">{eyebrow}</div>
      <h2 className="title serif">{title}</h2>
      {sub && <div className="sub">{sub}</div>}
    </div>
  )
}

function Stat({ n, unit, label }) {
  return (
    <div className="stat">
      <div className="stat-n display">
        {n}
        {unit && <span className="stat-unit">{unit}</span>}
      </div>
      <div className="stat-label">{label}</div>
    </div>
  )
}

function Footer() {
  return (
    <footer className="ft">
      <div className="container ft-row">
        <div className="ft-brand">
          <span className="brand-mark">星</span>
          <span>
            <strong className="display">星仔高端定制</strong>
            <small>工厂直营 · 全屋定制</small>
          </span>
        </div>
        <div className="ft-meta">
          <div>地址：广东省 佛山市 顺德区 龙江镇 国际家具产业基地 8 号</div>
          <div>热线：400-888-1314 · 邮箱：hello@xingzai.cn</div>
          <div className="muted">© 2026 星仔高端定制 · 粤 ICP 备 2026000001 号</div>
        </div>
      </div>
    </footer>
  )
}