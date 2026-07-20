import { useEffect, useMemo, useState } from 'react'
import { api } from '../../api'

const TYPE_LABEL = {
  style: '一级风格',
  space: '空间',
  color: '颜色',
  size: '尺寸',
  price: '价格'
}

export default function Tags() {
  const [all, setAll] = useState([])
  const [filter, setFilter] = useState('style')
  const [editing, setEditing] = useState(null)
  const [showForm, setShowForm] = useState(false)

  const load = () => api.adminListTags().then((r) => setAll(r.data || []))
  useEffect(() => { load() }, [])

  const grouped = useMemo(() => {
    const g = {}
    for (const t of all) {
      if (!g[t.type]) g[t.type] = []
      g[t.type].push(t)
    }
    return g
  }, [all])

  const onSave = async () => {
    if (editing.id) {
      await api.adminUpdateTag(editing.id, editing)
    } else {
      await api.adminCreateTag(editing)
    }
    setShowForm(false)
    setEditing(null)
    load()
  }

  const onDelete = async (id) => {
    if (!confirm('确认删除该标签?')) return
    await api.adminDeleteTag(id)
    load()
  }

  const startNew = () => {
    setEditing({ type: filter, name: '', value: '', color: '', icon: '', sort: 0, enabled: true })
    setShowForm(true)
  }
  const startEdit = (t) => {
    setEditing({ ...t })
    setShowForm(true)
  }

  const cur = grouped[filter] || []

  return (
    <div>
      <h1 className="adm-h1">标签管理</h1>
      <p className="adm-sub">一级风格 + 二级多维筛选标签</p>

      <div style={{ display: 'flex', gap: 10, marginBottom: 24, flexWrap: 'wrap' }}>
        {Object.keys(TYPE_LABEL).map((k) => (
          <button
            key={k}
            className={`chip ${filter === k ? 'active' : ''}`}
            onClick={() => setFilter(k)}
          >
            {TYPE_LABEL[k]} ({(grouped[k] || []).length})
          </button>
        ))}
      </div>

      <div className="adm-card">
        <div className="adm-card-h">
          <h2>{TYPE_LABEL[filter]} · {cur.length} 条</h2>
          <button className="btn btn-gold" onClick={startNew}>+ 新建标签</button>
        </div>
        <table className="adm-table">
          <thead>
            <tr>
              <th>名称</th><th>Value</th><th>色值</th><th>图标</th><th>排序</th><th>状态</th><th>操作</th>
            </tr>
          </thead>
          <tbody>
            {cur.map((t) => (
              <tr key={t.id}>
                <td><strong>{t.name}</strong></td>
                <td className="muted">{t.value}</td>
                <td>
                  {t.color ? (
                    <span style={{
                      display: 'inline-block', width: 28, height: 28, borderRadius: '50%',
                      background: t.color, outline: '1px solid var(--mist)'
                    }} />
                  ) : '-'}
                </td>
                <td className="muted">{t.icon}</td>
                <td>{t.sort}</td>
                <td className={t.enabled ? 'status-on' : 'status-off'}>{t.enabled ? '启用' : '禁用'}</td>
                <td className="adm-actions">
                  <button className="btn-mini primary" onClick={() => startEdit(t)}>编辑</button>
                  <button className="btn-mini danger" onClick={() => onDelete(t.id)}>删除</button>
                </td>
              </tr>
            ))}
            {cur.length === 0 && <tr><td colSpan={7} style={{ textAlign: 'center', padding: 40 }}>暂无数据</td></tr>}
          </tbody>
        </table>
      </div>

      {showForm && (
        <div className="adm-card">
          <div className="adm-card-h">
            <h2>{editing.id ? '编辑标签' : '新建标签'}</h2>
            <button className="btn-mini" onClick={() => setShowForm(false)}>取消</button>
          </div>
          <div className="adm-form">
            <label>类型
              <select value={editing.type} onChange={(e) => setEditing({ ...editing, type: e.target.value })}>
                {Object.entries(TYPE_LABEL).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
              </select>
            </label>
            <label>名称<input value={editing.name} onChange={(e) => setEditing({ ...editing, name: e.target.value })} /></label>
            <label>Value<input value={editing.value} onChange={(e) => setEditing({ ...editing, value: e.target.value })} /></label>
            <label>色值<input value={editing.color} onChange={(e) => setEditing({ ...editing, color: e.target.value })} placeholder="#FFFFFF" /></label>
            <label>图标<input value={editing.icon} onChange={(e) => setEditing({ ...editing, icon: e.target.value })} /></label>
            <label>排序<input type="number" value={editing.sort} onChange={(e) => setEditing({ ...editing, sort: +e.target.value })} /></label>
            <label>状态
              <select value={editing.enabled ? '1' : '0'} onChange={(e) => setEditing({ ...editing, enabled: e.target.value === '1' })}>
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </label>
            <div className="submit-row">
              <button className="btn-mini primary" onClick={onSave}>保存</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}