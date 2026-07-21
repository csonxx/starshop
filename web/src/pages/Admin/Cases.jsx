import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../../api'

export default function Cases() {
  const [list, setList] = useState([])
  const [filterStyle, setFilterStyle] = useState('')
  const [styles, setStyles] = useState([])
  const [error, setError] = useState('')
  const nav = useNavigate()

  const load = () => api.adminListCases().then((r) => setList(r.data || [])).catch((err) => setError(err.message))
  useEffect(() => {
    load()
    api.tags('style').then((r) => setStyles(r.data || [])).catch((err) => setError(err.message))
  }, [])

  const filtered = filterStyle
    ? list.filter((c) => c.style === filterStyle)
    : list

  const onDelete = async (id) => {
    if (!confirm('确认删除该案例?')) return
    try {
      await api.adminDeleteCase(id)
      await load()
    } catch (err) {
      setError(err.message)
    }
  }
  const onTogglePin = async (id) => {
    try {
      await api.adminTogglePin(id)
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  return (
    <div>
      <h1 className="adm-h1">案例管理</h1>
      <p className="adm-sub">新增 / 编辑 / 上下架 / 置顶 · 数据源：MongoDB</p>
      {error && <div className="adm-message error">{error}</div>}

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24, flexWrap: 'wrap', gap: 12 }}>
        <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap' }}>
          <button className={`chip ${!filterStyle ? 'active' : ''}`} onClick={() => setFilterStyle('')}>全部 ({list.length})</button>
          {styles.map((s) => (
            <button
              key={s.id}
              className={`chip ${filterStyle === s.value ? 'active' : ''}`}
              onClick={() => setFilterStyle(s.value)}
            >
              {s.name} ({list.filter(c => c.style === s.value).length})
            </button>
          ))}
        </div>
        <Link to="/admin/cases/new" className="btn btn-gold">+ 新增案例</Link>
      </div>

      <div className="adm-card">
        <table className="adm-table">
          <thead>
            <tr>
              <th>封面</th><th>标题</th><th>风格</th><th>空间</th><th>价格</th><th>置顶</th><th>状态</th><th>操作</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((c) => (
              <tr key={c.id}>
                <td><img className="thumb" src={c.cover} alt="" /></td>
                <td>
                  <strong>{c.title}</strong>
                  <div className="muted" style={{ fontSize: 11 }}>{c.size} · {c.area}</div>
                </td>
                <td><span className="status-on">{c.style}</span></td>
                <td>{c.space}</td>
                <td><strong className="display">¥{(c.price / 10000).toFixed(2)}万</strong></td>
                <td>
                  {c.pinned
                    ? <span className="status-on">🔥 置顶</span>
                    : <span className="muted">—</span>}
                </td>
                <td className={c.enabled ? 'status-on' : 'status-off'}>{c.enabled ? '上架' : '下架'}</td>
                <td className="adm-actions">
                  <button className="btn-mini primary" onClick={() => nav(`/admin/cases/${c.id}`)}>编辑</button>
                  <button className="btn-mini gold" onClick={() => onTogglePin(c.id)}>{c.pinned ? '取消置顶' : '置顶'}</button>
                  <button className="btn-mini danger" onClick={() => onDelete(c.id)}>删除</button>
                </td>
              </tr>
            ))}
            {filtered.length === 0 && <tr><td colSpan={8} style={{ textAlign: 'center', padding: 40 }}>暂无案例</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}
