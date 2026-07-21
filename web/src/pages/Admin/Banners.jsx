import { useEffect, useState } from 'react'
import { api } from '../../api'

export default function Banners() {
  const [list, setList] = useState([])
  const [editing, setEditing] = useState(null)
  const [showForm, setShowForm] = useState(false)
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  const load = () => api.adminListBanners().then((r) => setList(r.data || [])).catch((err) => setError(err.message))
  useEffect(() => { load() }, [])

  const onSave = async () => {
    setSaving(true)
    setError('')
    try {
      if (editing.id) {
        await api.adminUpdateBanner(editing.id, editing)
      } else {
        await api.adminCreateBanner(editing)
      }
      setShowForm(false)
      setEditing(null)
      await load()
    } catch (err) {
      setError(err.message)
    } finally {
      setSaving(false)
    }
  }

  const onDelete = async (id) => {
    if (!confirm('确认删除该 Banner?')) return
    try {
      await api.adminDeleteBanner(id)
      await load()
    } catch (err) {
      setError(err.message)
    }
  }

  const startNew = () => {
    setEditing({ title: '', subtitle: '', image: '', link: '', sort: 0, enabled: true })
    setShowForm(true)
  }
  const startEdit = (b) => {
    setEditing({ ...b })
    setShowForm(true)
  }

  return (
    <div>
      <h1 className="adm-h1">Banner 管理</h1>
      <p className="adm-sub">运营首页顶部轮播 · 最多 5 张</p>
      {error && <div className="adm-message error">{error}</div>}

      <div className="adm-card">
        <div className="adm-card-h">
          <h2>Banner 列表</h2>
          <button className="btn btn-gold" onClick={startNew}>+ 新建 Banner</button>
        </div>
        <table className="adm-table">
          <thead>
            <tr>
              <th>封面</th><th>标题</th><th>副标题</th><th>链接</th><th>排序</th><th>状态</th><th>操作</th>
            </tr>
          </thead>
          <tbody>
            {list.map((b) => (
              <tr key={b.id}>
                <td><img className="thumb" src={b.image} alt="" /></td>
                <td><strong>{b.title}</strong></td>
                <td className="muted">{b.subtitle}</td>
                <td className="muted">{b.link}</td>
                <td>{b.sort}</td>
                <td className={b.enabled ? 'status-on' : 'status-off'}>{b.enabled ? '启用' : '禁用'}</td>
                <td className="adm-actions">
                  <button className="btn-mini primary" onClick={() => startEdit(b)}>编辑</button>
                  <button className="btn-mini danger" onClick={() => onDelete(b.id)}>删除</button>
                </td>
              </tr>
            ))}
            {list.length === 0 && <tr><td colSpan={7} style={{ textAlign: 'center', padding: 40 }}>暂无 Banner</td></tr>}
          </tbody>
        </table>
      </div>

      {showForm && (
        <div className="adm-card">
          <div className="adm-card-h">
            <h2>{editing.id ? '编辑 Banner' : '新建 Banner'}</h2>
            <button className="btn-mini" onClick={() => setShowForm(false)}>取消</button>
          </div>
          <div className="adm-form">
            <label className="full">标题<input value={editing.title} onChange={(e) => setEditing({ ...editing, title: e.target.value })} /></label>
            <label className="full">副标题<input value={editing.subtitle} onChange={(e) => setEditing({ ...editing, subtitle: e.target.value })} /></label>
            <label className="full">图片 URL<input value={editing.image} onChange={(e) => setEditing({ ...editing, image: e.target.value })} placeholder="https://..." /></label>
            <label className="full">跳转链接<input value={editing.link} onChange={(e) => setEditing({ ...editing, link: e.target.value })} placeholder="/cases 或案例 id" /></label>
            <label>排序<input type="number" value={editing.sort} onChange={(e) => setEditing({ ...editing, sort: +e.target.value })} /></label>
            <label>状态
              <select value={editing.enabled ? '1' : '0'} onChange={(e) => setEditing({ ...editing, enabled: e.target.value === '1' })}>
                <option value="1">启用</option>
                <option value="0">禁用</option>
              </select>
            </label>
            <div className="submit-row">
              <button className="btn-mini primary" onClick={onSave} disabled={saving}>{saving ? '保存中...' : '保存'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
