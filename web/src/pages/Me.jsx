import { Link } from 'react-router-dom'
import { useUser } from '../store/user.jsx'
import './Me.css'

export default function Me() {
  const { user, isAdmin, logout } = useUser()
  if (!user) {
    return (
      <div className="me container">
        <p>请先 <Link to="/login">登录</Link></p>
      </div>
    )
  }
  return (
    <div className="me container">
      <div className="me-card">
        <div className="me-avatar">{user.phone?.slice(-2)}</div>
        <h2>{user.nickname || 'Star 用户'}</h2>
        <p className="me-phone">{user.phone}</p>
        <p className="me-role">身份 · {user.role === 'admin' ? '管理员' : '普通用户'}</p>
        {isAdmin && <Link to="/admin" className="btn btn-gold">进入运营后台</Link>}
        <button className="btn btn-ghost" onClick={logout}>退出登录</button>
      </div>
    </div>
  )
}