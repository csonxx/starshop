import { useEffect, useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { api } from '../api'
import CaseCard from '../components/CaseCard'
import './StylePage.css'

// 不同空间对应的可用尺寸集合 —— 空间变化时尺寸联动
const SPACE_SIZES = {
  '客厅':     ['2.0m', '2.4m', '通顶'],
  '餐厅':     ['1.5m', '1.8m', '2.0m', '2.4m'],
  '主卧':     ['1.8m', '2.0m', '2.4m', '通顶'],
  '次卧':     ['1.2m', '1.5m', '1.8m', '2.0m'],
  '书房':     ['1.2m', '1.5m', '1.8m', '2.0m', '通顶'],
  '衣帽间':   ['2.0m', '2.4m', '通顶'],
  '玄关':     ['1.2m', '1.5m'],
  '儿童房':   ['1.2m', '1.5m', '1.8m'],
  '多功能房': ['1.5m', '1.8m', '2.0m', '通顶']
}

export default function StylePage() {
  const { slug } = useParams()
  const [params, setParams] = useSearchParams()

  const [styleTags, setStyleTags] = useState([])
  const [secondary, setSecondary] = useState({
    space: [], color: [], size: [], price: []
  })

  const [activeStyle, setActiveStyle] = useState(slug || '')
  const [filter, setFilter] = useState({
    space: params.get('space') || '',
    color: params.get('color') || '',
    size:  params.get('size')  || '',
    price: params.get('price') || ''
  })
  const [cases, setCases] = useState([])
  const [loading, setLoading] = useState(true)

  // 加载所有标签 + 风格列表
  useEffect(() => {
    Promise.all([
      api.tags('style'),
      api.tags('space'),
      api.tags('color'),
      api.tags('size'),
      api.tags('price')
    ]).then(([st, sp, cl, sz, pr]) => {
      setStyleTags(st.data || [])
      setSecondary({
        space: sp.data || [],
        color: cl.data || [],
        size: sz.data || [],
        price: pr.data || []
      })
      setLoading(false)
    }).catch((e) => { console.error(e); setLoading(false) })
  }, [])

  // 路由 slug 切换 → 重置
  useEffect(() => {
    setActiveStyle(slug || '')
  }, [slug])

  // 空间变化时，尺寸选项随之变化（若当前尺寸不合法则清空）
  useEffect(() => {
    if (!filter.space) return
    const allowed = SPACE_SIZES[filter.space] || []
    if (filter.size && !allowed.includes(filter.size)) {
      setFilter((cur) => ({ ...cur, size: '' }))
    }
  }, [filter.space])

  // 查询案例
  useEffect(() => {
    if (loading) return
    const q = { page: 1, pageSize: 60 }
    if (activeStyle) q.style = activeStyle
    if (filter.space) q.space = filter.space
    if (filter.color) q.color = filter.color
    if (filter.size) q.size = filter.size
    if (filter.price) q.price = filter.price
    api.cases(q).then((r) => setCases(r.data.list || []))
      .catch(console.error)
  }, [activeStyle, filter, loading])

  const availSizes = filter.space
    ? (SPACE_SIZES[filter.space] || [])
    : secondary.size.map((t) => t.value)

  const onStylePick = (val) => {
    if (val === activeStyle) return
    window.location.href = `/style/${val}`
  }

  const onSecPick = (key, val) => {
    setFilter((cur) => ({ ...cur, [key]: cur[key] === val ? '' : val }))
  }

  const onClearAll = () => {
    setFilter({ space: '', color: '', size: '', price: '' })
  }

  const currentStyle = styleTags.find((s) => s.value === activeStyle)

  return (
    <div className="style-page">
      <section className="sp-hero">
        <div className="container">
          <div className="sp-eyebrow display">STYLE · {activeStyle?.toUpperCase() || 'ALL'}</div>
          <h1 className="sp-title serif">
            {currentStyle
              ? <>甄选 <span className="hl">{currentStyle.name}</span> 全屋定制案例</>
              : <>甄选全部主流风格 · 全屋定制案例</>}
          </h1>
          <p className="sp-sub">{cases.length} 个作品 · 工厂直营 · 设计师 1V1</p>
        </div>
      </section>

      <section className="sp-chips">
        <div className="container">
          <div className="sp-eyebrow-row">
            <span className="display">01 · 风格 STYLE</span>
            <span className="muted">切换风格跳转到对应筛选页</span>
          </div>
          <div className="sp-style-chips">
            <Link
              to="/cases"
              className={`chip ${!activeStyle ? 'active' : ''}`}
            >全部</Link>
            {styleTags.map((t) => (
              <Link
                key={t.id}
                to={`/style/${t.value}`}
                className={`chip ${activeStyle === t.value ? 'active' : ''}`}
              >{t.name}</Link>
            ))}
          </div>
        </div>
      </section>

      <section className="sp-filter">
        <div className="container">
          <div className="sp-eyebrow-row">
            <span className="display">02 · 筛选 FILTER</span>
            <span className="muted">
              空间 · 颜色 · 尺寸 · 价格
              {filter.space && ` · 当前空间「${filter.space}」`}
            </span>
            {(filter.space || filter.color || filter.size || filter.price) && (
              <button className="sp-clear" onClick={onClearAll}>清除筛选</button>
            )}
          </div>

          <div className="sp-filter-card">
            <FilterRow label="空间">
              <button
                className={`chip ${!filter.space ? 'active' : ''}`}
                onClick={() => onSecPick('space', '')}
              >不限</button>
              {secondary.space.map((t) => (
                <button
                  key={t.id}
                  className={`chip ${filter.space === t.value ? 'active' : ''}`}
                  onClick={() => onSecPick('space', t.value)}
                >{t.name}</button>
              ))}
            </FilterRow>

            <FilterRow label="颜色">
              <button
                className={`color-dot ${!filter.color ? 'active' : ''}`}
                style={{ background: 'transparent', outline: '1px dashed var(--mist)' }}
                onClick={() => onSecPick('color', '')}
                data-name="不限"
              />
              {secondary.color.map((t) => (
                <button
                  key={t.id}
                  className={`color-dot ${filter.color === t.value ? 'active' : ''}`}
                  style={{ background: t.color }}
                  data-name={t.name}
                  onClick={() => onSecPick('color', t.value)}
                  title={t.name}
                />
              ))}
            </FilterRow>

            <FilterRow label={filter.space ? `尺寸 (依「${filter.space}」联动)` : '尺寸'}>
              <button
                className={`chip ${!filter.size ? 'active' : ''}`}
                onClick={() => onSecPick('size', '')}
              >不限</button>
              {secondary.size
                .filter((t) => filter.space === '' || availSizes.includes(t.value))
                .map((t) => (
                  <button
                    key={t.id}
                    className={`chip ${filter.size === t.value ? 'active' : ''}`}
                    onClick={() => onSecPick('size', t.value)}
                  >
                    {t.name}{t.value !== '通顶' ? ' 书柜' : t.value === '通顶' && filter.space && SPACE_SIZES[filter.space]?.includes('通顶') ? '' : ''}
                  </button>
                ))}
            </FilterRow>

            <FilterRow label="价格">
              <button
                className={`chip ${!filter.price ? 'active' : ''}`}
                onClick={() => onSecPick('price', '')}
              >不限</button>
              {secondary.price.map((t) => (
                <button
                  key={t.id}
                  className={`chip ${filter.price === t.value ? 'active' : ''}`}
                  onClick={() => onSecPick('price', t.value)}
                >{t.name}</button>
              ))}
            </FilterRow>
          </div>
        </div>
      </section>

      <section className="sp-list">
        <div className="container">
          <div className="sp-eyebrow-row">
            <span className="display">03 · 案例 WORKS</span>
            <span className="muted">
              {cases.length} 个作品
              {activeStyle && currentStyle && ` · ${currentStyle.name}`}
            </span>
          </div>

          {loading && <div className="empty">载入中...</div>}
          {!loading && cases.length === 0 && (
            <div className="empty">暂无匹配案例 · 试试切换风格或筛选条件</div>
          )}
          <div className="sp-grid">
            {cases.map((c) => <CaseCard key={c.id} item={c} />)}
          </div>
        </div>
      </section>

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
    </div>
  )
}

function FilterRow({ label, children }) {
  return (
    <div className="sp-filter-row">
      <div className="sp-filter-label">{label}</div>
      <div className="sp-filter-content">{children}</div>
    </div>
  )
}