import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { useEffect } from 'react'
import { useUser } from '../../store/user.jsx'
import './Admin.css'

export default function AdminLayout() {
  const { user, isAdmin, logout } = useUser()
  const nav = useNavigate()

  useEffect(() => {
    if (!user) {
      nav('/login?redirect=/admin', { replace: true })
    } else if (!isAdmin) {
      nav('/', { replace: true })
    }
  }, [user, isAdmin, nav])

  if (!user || !isAdmin) return null

  return (
    <div className="admin">
      <aside className="adm-side">
        <div className="adm-brand">
          <span className="brand-mark">星</span>
          <span>
            <strong className="display">星仔运营后台</strong>
            <small>运营后台 · v1.0</small>
          </span>
        </div>
        <nav className="adm-nav">
          <NavLink to="/admin" end>📊 数据概览</NavLink>
          <NavLink to="/admin/banners">🖼  Banner 管理</NavLink>
          <NavLink to="/admin/tags">🏷  标签管理</NavLink>
          <NavLink to="/admin/cases">📦  案例管理</NavLink>
          <NavLink to="/admin/cases/new">➕  新增案例</NavLink>
        </nav>
        <div className="adm-foot">
          <div className="adm-user">{user.phone}</div>
          <button className="adm-logout" onClick={() => { logout(); nav('/') }}>退出</button>
        </div>
      </aside>
      <main className="adm-main">
        <Outlet />
      </main>
    </div>
  )
}