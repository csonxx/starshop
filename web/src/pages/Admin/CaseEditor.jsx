import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api } from '../../api'

const empty = {
  title: '',
  style: 'new-chinese',
  space: '客厅',
  colors: [],
  size: '2.0m',
  area: '',
  price: 0,
  priceLabel: '1-3万',
  cover: '',
  images: [],
  highlights: [],
  materials: [],
  hardware: [],
  pinned: false,
  enabled: true
}

export default function CaseEditor() {
  const { id } = useParams()
  const isNew = !id || id === 'new'
  const nav = useNavigate()
  const [data, setData] = useState(empty)
  const [styles, setStyles] = useState([])
  const [spaces, setSpaces] = useState([])
  const [sizes, setSizes] = useState([])
  const [prices, setPrices] = useState([])
  const [colors, setColors] = useState([])

  useEffect(() => {
    Promise.all([
      api.tags('style'),
      api.tags('space'),
      api.tags('size'),
      api.tags('price'),
      api.tags('color')
    ]).then(([a, b, c, d, e]) => {
      setStyles(a.data || [])
      setSpaces(b.data || [])
      setSizes(c.data || [])
      setPrices(d.data || [])
      setColors(e.data || [])
    })
    if (!isNew) {
      api.caseDetail(id).then((r) => setData(r.data)).catch(() => setData(empty))
    }
  }, [id])

  const onSubmit = async () => {
    const payload = {
      ...data,
      price: Number(data.price) || 0,
      colors: arrayField(data.colors),
      images: arrayField(data.images),
      highlights: arrayField(data.highlights),
      materials: arrayField(data.materials),
      hardware: arrayField(data.hardware)
    }
    if (isNew) {
      await api.adminCreateCase(payload)
    } else {
      await api.adminUpdateCase(id, payload)
    }
    nav('/admin/cases')
  }

  return (
    <div>
      <h1 className="adm-h1">{isNew ? '新增案例' : '编辑案例'}</h1>
      <p className="adm-sub">填写案例详情 · 支持多色 / 多图 / 多亮点</p>

      <div className="adm-form">
        <label className="full">标题<input value={data.title} onChange={(e) => setData({ ...data, title: e.target.value })} placeholder="例如：云山胡桃 · 通顶玄关一体柜" /></label>

        <label>风格
          <select value={data.style} onChange={(e) => setData({ ...data, style: e.target.value })}>
            {styles.map((s) => <option key={s.id} value={s.value}>{s.name}</option>)}
          </select>
        </label>
        <label>空间
          <select value={data.space} onChange={(e) => setData({ ...data, space: e.target.value })}>
            {spaces.map((s) => <option key={s.id} value={s.value}>{s.name}</option>)}
          </select>
        </label>
        <label>尺寸
          <select value={data.size} onChange={(e) => setData({ ...data, size: e.target.value })}>
            {sizes.map((s) => <option key={s.id} value={s.value}>{s.name}</option>)}
          </select>
        </label>
        <label>面积<input value={data.area} onChange={(e) => setData({ ...data, area: e.target.value })} placeholder="例如 13㎡" /></label>
        <label>价格 (元)<input type="number" value={data.price} onChange={(e) => setData({ ...data, price: e.target.value })} /></label>
        <label>价格区间
          <select value={data.priceLabel} onChange={(e) => setData({ ...data, priceLabel: e.target.value })}>
            {prices.map((s) => <option key={s.id} value={s.value}>{s.name}</option>)}
          </select>
        </label>

        <label className="full">封面图 URL<input value={data.cover} onChange={(e) => setData({ ...data, cover: e.target.value })} placeholder="https://..." /></label>
        <label className="full">多图 URL (换行分隔)<textarea value={(data.images || []).join('\n')} onChange={(e) => setData({ ...data, images: e.target.value.split('\n') })} /></label>

        <label className="full">颜色 (可多选 · 来自颜色标签)
          <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', padding: '8px 0' }}>
            {colors.map((c) => {
              const active = (data.colors || []).includes(c.value)
              return (
                <button
                  type="button"
                  key={c.id}
                  className={`color-dot ${active ? 'active' : ''}`}
                  style={{ background: c.color }}
                  data-name={c.name}
                  onClick={() => {
                    const cur = new Set(data.colors || [])
                    if (cur.has(c.value)) cur.delete(c.value); else cur.add(c.value)
                    setData({ ...data, colors: [...cur] })
                  }}
                />
              )
            })}
          </div>
        </label>

        <label className="full">设计亮点 (每行一条)<textarea value={(data.highlights || []).join('\n')} onChange={(e) => setData({ ...data, highlights: e.target.value.split('\n') })} /></label>
        <label className="full">主材 (每行一条)<textarea value={(data.materials || []).join('\n')} onChange={(e) => setData({ ...data, materials: e.target.value.split('\n') })} /></label>
        <label className="full">五金 (每行一条)<textarea value={(data.hardware || []).join('\n')} onChange={(e) => setData({ ...data, hardware: e.target.value.split('\n') })} /></label>

        <label>置顶
          <select value={data.pinned ? '1' : '0'} onChange={(e) => setData({ ...data, pinned: e.target.value === '1' })}>
            <option value="0">普通</option>
            <option value="1">🔥 置顶</option>
          </select>
        </label>
        <label>状态
          <select value={data.enabled ? '1' : '0'} onChange={(e) => setData({ ...data, enabled: e.target.value === '1' })}>
            <option value="1">上架</option>
            <option value="0">下架</option>
          </select>
        </label>

        <div className="submit-row">
          <button className="btn-mini" onClick={() => nav('/admin/cases')}>取消</button>
          <button className="btn-mini primary" onClick={onSubmit}>保存</button>
        </div>
      </div>
    </div>
  )
}

function arrayField(v) {
  if (Array.isArray(v)) return v.filter(Boolean)
  return String(v || '').split('\n').map((x) => x.trim()).filter(Boolean)
}