import { Link, NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useUser } from '../store/user.jsx'
import { useEffect, useState } from 'react'
import './Header.css'

export default function Header() {
  const { user, isAdmin, logout } = useUser()
  const nav = useNavigate()
  const loc = useLocation()
  const [open, setOpen] = useState(false)

  useEffect(() => { setOpen(false) }, [loc.pathname])

  return (
    <header className={`hdr ${open ? 'open' : ''}`}>
      <div className="container hdr-row">
        <Link to="/" className="brand">
          <span className="brand-mark">星</span>
          <span className="brand-name">
            <em className="display">星仔</em>
            <small>ATELIER · 工厂直营</small>
          </span>
        </Link>

        <nav className="nav">
          <NavLink to="/" end>首页</NavLink>
          <NavLink to="/style/new-chinese">新中式</NavLink>
          <NavLink to="/style/cream">奶油风</NavLink>
          <NavLink to="/style/italian-luxury">意式轻奢</NavLink>
          <NavLink to="/style/modern">现代简约</NavLink>
          <NavLink to="/cases" end>全部分类</NavLink>
          <a href="#factory">工厂直营</a>
        </nav>

        <div className="hdr-right">
          {user ? (
            <>
              {isAdmin && <Link to="/admin" className="admin-link">运营后台</Link>}
              <button className="user-chip" onClick={() => nav('/me')}>
                <span>{user.phone}</span>
              </button>
              <button className="logout" onClick={() => { logout(); nav('/') }}>退出</button>
            </>
          ) : (
            <Link to="/login" className="btn btn-ghost">登录 / 注册</Link>
          )}
          <button className="hamburger" onClick={() => setOpen(!open)} aria-label="菜单">
            <span /><span /><span />
          </button>
        </div>
      </div>

      {open && (
        <div className="mobile-menu">
          <NavLink to="/" end>首页</NavLink>
          <NavLink to="/style/new-chinese">新中式</NavLink>
          <NavLink to="/style/cream">奶油风</NavLink>
          <NavLink to="/style/italian-luxury">意式轻奢</NavLink>
          <NavLink to="/style/modern">现代简约</NavLink>
          <NavLink to="/cases" end>全部分类</NavLink>
          <a href="#factory" onClick={() => setOpen(false)}>工厂直营</a>
          {user ? (
            isAdmin && <Link to="/admin">运营后台</Link>
          ) : (
            <Link to="/login">登录 / 注册</Link>
          )}
        </div>
      )}
    </header>
  )
}