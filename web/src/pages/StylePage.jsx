import { useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { api } from '../api'
import CaseCard from '../components/CaseCard'
import './StylePage.css'

// 不同空间对应的可用尺寸集合 —— 空间变化时尺寸联动
//
// 真实业务规格分类 (对应后端 SPACE_SIZES_MAP):
//   主卧/次卧/衣帽间  - 衣柜深度 500/550/560/580 + 高度 2.4 / 通顶 / U 型
//   客厅              - 电视柜 2.0m / 2.4m / 3.0m / 满墙
//   餐厅              - 餐边柜 1.2 / 1.5 / 1.8 岛台 / 2.0 半高
//   书房              - 书桌 1.2 / 1.6 / 榻榻米 1500/1800
//   玄关              - 鞋柜 1.0 / 1.2 通顶 / 1.5 / 换鞋凳
//   儿童房            - 子母床/上下铺/书桌+衣柜/高低床
//   多功能房          - 与书房相近
const SPACE_SIZES = {
  '主卧':   ['560深·2.4m高', '560深·2.7m通顶', '580深·一门到顶', '衣帽间U型'],
  '次卧':   ['500深·2.4m高', '550深·2.4m高', '560深·2.4m高', '2.4m顶柜+挂衣'],
  '衣帽间': ['560深·2.4m高', '560深·2.7m通顶', 'U型步入式', 'L型步入式'],
  '客厅':   ['2.0m悬空电视柜', '2.4m满墙电视柜', '3.0m展示柜', '一字到顶收纳柜'],
  '餐厅':   ['1.2m餐边柜', '1.5m餐边柜', '1.8m岛台一体', '2.0m半高餐边'],
  '书房':   ['1.2m书桌', '1.6m书桌', '1500×2000榻榻米', '1800×2000升降桌'],
  '玄关':   ['1.0m三门鞋柜', '1.2m通顶鞋柜', '1.5m到顶鞋柜', '换鞋凳一体'],
  '儿童房': ['子母床·1.5m', '上下铺·1.8m', '1.2m书桌+衣柜', '成长型高低床'],
  '多功能房': ['1500×2000榻榻米', '1800×2000升降桌', '1.6m书桌', '一字到顶收纳柜']
}

const FILTER_KEYS = ['space', 'color', 'size', 'price']
const PAGE_SIZE = 24

function readFilter(params) {
  return FILTER_KEYS.reduce((result, key) => {
    result[key] = params.get(key)?.split(',').map((value) => value.trim()).filter(Boolean) || []
    return result
  }, {})
}

function filterSearch(filter) {
  const search = new URLSearchParams()
  FILTER_KEYS.forEach((key) => {
    if (filter[key].length) search.set(key, filter[key].join(','))
  })
  return search
}

function sameFilter(left, right) {
  return FILTER_KEYS.every((key) => left[key].join(',') === right[key].join(','))
}

export default function StylePage() {
  const { slug } = useParams()
  const [params, setParams] = useSearchParams()

  const [styleTags, setStyleTags] = useState([])
  const [secondary, setSecondary] = useState({
    space: [], color: [], size: [], price: []
  })

  const activeStyle = slug || ''
  const [filter, setFilter] = useState(() => readFilter(params))
  const [cases, setCases] = useState([])
  const [total, setTotal] = useState(0)
  const [tagsLoading, setTagsLoading] = useState(true)
  const [casesLoading, setCasesLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [tagsError, setTagsError] = useState('')
  const [casesError, setCasesError] = useState('')
  const requestRef = useRef(0)

  // 加载所有标签 + 风格列表
  useEffect(() => {
    const controller = new AbortController()
    Promise.all([
      api.tags('style', { signal: controller.signal }),
      api.tags('space', { signal: controller.signal }),
      api.tags('color', { signal: controller.signal }),
      api.tags('size', { signal: controller.signal }),
      api.tags('price', { signal: controller.signal })
    ]).then(([st, sp, cl, sz, pr]) => {
      setStyleTags(st.data || [])
      setSecondary({
        space: sp.data || [],
        color: cl.data || [],
        size: sz.data || [],
        price: pr.data || []
      })
    }).catch((e) => {
      if (e.name !== 'CanceledError') setTagsError(e.message || '筛选项加载失败')
    }).finally(() => {
      if (!controller.signal.aborted) setTagsLoading(false)
    })
    return () => controller.abort()
  }, [])

  useEffect(() => {
    const next = readFilter(params)
    setFilter((current) => sameFilter(current, next) ? current : next)
  }, [params])

  // 空间变化时，尺寸选项随之变化（若当前尺寸不合法则清空）
  useEffect(() => {
    if (filter.space.length !== 1) return
    const allowed = SPACE_SIZES[filter.space[0]] || []
    const sizes = filter.size.filter((value) => allowed.includes(value))
    if (sizes.length !== filter.size.length) {
      updateFilter({ ...filter, size: sizes })
    }
  }, [filter.space, filter.size])

  useEffect(() => {
    const controller = new AbortController()
    const requestID = ++requestRef.current
    const q = { page: 1, pageSize: PAGE_SIZE }
    if (activeStyle) q.style = activeStyle
    FILTER_KEYS.forEach((key) => {
      if (filter[key].length) q[key] = filter[key].join(',')
    })
    setCasesLoading(true)
    setLoadingMore(false)
    setCasesError('')
    api.cases(q, { signal: controller.signal }).then((r) => {
      if (requestID !== requestRef.current) return
      setCases(r.data.list || [])
      setTotal(r.data.total || 0)
    }).catch((e) => {
      if (e.name !== 'CanceledError' && requestID === requestRef.current) {
        setCases([])
        setTotal(0)
        setCasesError(e.message || '案例加载失败')
      }
    }).finally(() => {
      if (requestID === requestRef.current) setCasesLoading(false)
    })
    return () => controller.abort()
  }, [activeStyle, filter])

  // 当选了空间时, 尺寸选项是 SPACE_SIZES[space] (业务规格)
  // 当没选空间时, 给所有真实业务规格去重展示
  const availSizes = useMemo(() => {
    if (filter.space.length === 1) {
      return SPACE_SIZES[filter.space[0]] || []
    }
    const all = new Set()
    Object.values(SPACE_SIZES).forEach((arr) => arr.forEach((s) => all.add(s)))
    return Array.from(all)
  }, [filter.space])

  const onSecPick = (key, val) => {
    if (!val) {
      updateFilter({ ...filter, [key]: [] })
      return
    }
    const current = filter[key]
    const next = current.includes(val)
      ? current.filter((value) => value !== val)
      : [...current, val]
    updateFilter({ ...filter, [key]: next })
  }

  const onClearAll = () => {
    updateFilter({ space: [], color: [], size: [], price: [] })
  }

  const updateFilter = (next) => {
    requestRef.current += 1
    setCasesLoading(true)
    setLoadingMore(false)
    setParams(filterSearch(next), { replace: true })
  }

  const loadMore = async () => {
    if (loadingMore || casesLoading || cases.length >= total) return
    const requestID = requestRef.current
    const q = { page: Math.floor(cases.length / PAGE_SIZE) + 1, pageSize: PAGE_SIZE }
    if (activeStyle) q.style = activeStyle
    FILTER_KEYS.forEach((key) => {
      if (filter[key].length) q[key] = filter[key].join(',')
    })
    setLoadingMore(true)
    setCasesError('')
    try {
      const r = await api.cases(q)
      if (requestID !== requestRef.current) return
      setCases((current) => {
        const seen = new Set(current.map((item) => item.id))
        return [...current, ...(r.data.list || []).filter((item) => !seen.has(item.id))]
      })
      setTotal(r.data.total || 0)
    } catch (e) {
      if (requestID === requestRef.current) setCasesError(e.message || '更多案例加载失败')
    } finally {
      if (requestID === requestRef.current) setLoadingMore(false)
    }
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
          <p className="sp-sub">{total} 个作品 · 工厂直营 · 设计师 1V1</p>
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
              to={{ pathname: '/cases', search: params.toString() }}
              className={`chip ${!activeStyle ? 'active' : ''}`}
            >全部</Link>
            {styleTags.map((t) => (
              <Link
                key={t.id}
                to={{ pathname: `/style/${t.value}`, search: params.toString() }}
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
              {filter.space.length > 0 && ` · 已选 ${filter.space.length} 个空间`}
            </span>
            {FILTER_KEYS.some((key) => filter[key].length > 0) && (
              <button className="sp-clear" onClick={onClearAll}>清除筛选</button>
            )}
          </div>

          <div className="sp-filter-card">
            <FilterRow label="空间">
              <button
                className={`chip ${filter.space.length === 0 ? 'active' : ''}`}
                onClick={() => onSecPick('space', '')}
              >不限</button>
              {secondary.space.map((t) => (
                <button
                  key={t.id}
                  className={`chip ${filter.space.includes(t.value) ? 'active' : ''}`}
                  onClick={() => onSecPick('space', t.value)}
                >{t.name}</button>
              ))}
            </FilterRow>

            <FilterRow label="颜色">
              <button
                className={`color-dot ${filter.color.length === 0 ? 'active' : ''}`}
                style={{ background: 'transparent', outline: '1px dashed var(--mist)' }}
                onClick={() => onSecPick('color', '')}
                data-name="不限"
              />
              {secondary.color.map((t) => (
                <button
                  key={t.id}
                  className={`color-dot ${filter.color.includes(t.value) ? 'active' : ''}`}
                  style={{ background: t.color }}
                  data-name={t.name}
                  onClick={() => onSecPick('color', t.value)}
                  title={t.name}
                />
              ))}
            </FilterRow>

            <FilterRow label={filter.space.length === 1 ? `尺寸 (依「${filter.space[0]}」联动)` : '尺寸'}>
              <button
                className={`chip ${filter.size.length === 0 ? 'active' : ''}`}
                onClick={() => onSecPick('size', '')}
              >不限</button>
              {availSizes.map((s) => (
                <button
                  key={s}
                  className={`chip ${filter.size.includes(s) ? 'active' : ''}`}
                  onClick={() => onSecPick('size', s)}
                >{s}</button>
              ))}
            </FilterRow>

            <FilterRow label="价格">
              <button
                className={`chip ${filter.price.length === 0 ? 'active' : ''}`}
                onClick={() => onSecPick('price', '')}
              >不限</button>
              {secondary.price.map((t) => (
                <button
                  key={t.id}
                  className={`chip ${filter.price.includes(t.value) ? 'active' : ''}`}
                  onClick={() => onSecPick('price', t.value)}
                >{t.name}</button>
              ))}
            </FilterRow>
          </div>
        </div>
      </section>

      <section className="sp-list" id="cases">
        <div className="container">
          <div className="sp-eyebrow-row">
            <span className="display">03 · 案例 WORKS</span>
            <span className="muted">
              {total} 个作品
              {activeStyle && currentStyle && ` · ${currentStyle.name}`}
            </span>
          </div>

          {tagsLoading && casesLoading && <div className="empty">载入中...</div>}
          {(tagsError || casesError) && (
            <div className="sp-error" role="alert">
              <span>{tagsError || casesError}</span>
              <button className="btn btn-ghost" onClick={() => window.location.reload()}>重新加载</button>
            </div>
          )}
          {!casesLoading && !tagsError && !casesError && cases.length === 0 && (
            <div className="empty">暂无匹配案例 · 试试切换风格或筛选条件</div>
          )}
          <div className="sp-grid">
            {cases.map((c) => <CaseCard key={c.id} item={c} />)}
          </div>
          {cases.length > 0 && cases.length < total && (
            <div className="sp-load-more">
              <button className="btn btn-primary" onClick={loadMore} disabled={loadingMore}>
                {loadingMore ? '加载中...' : `加载更多 · 还有 ${total - cases.length} 个`}
              </button>
            </div>
          )}
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
