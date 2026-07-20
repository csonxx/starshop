// store/user.jsx - 全局用户 Context
//
// 职责:
//   1. 把当前用户资料持久化到 localStorage (登录态保留)
//   2. 把 JWT token 存到 localStorage (被 http.js 自动加到 Authorization 头)
//   3. 提供 isAdmin 判断, 供 CaseCard 等组件显示管理提示
//
// 用法:
//   const { user, isAdmin, login, logout } = useUser()
import { createContext, useContext, useEffect, useMemo, useState } from 'react'
import { api } from '../api'

const UserCtx = createContext(null)

export function UserProvider({ children }) {
  // 初始: 从 localStorage 反序列化 (用户关闭浏览器再次打开仍然登录)
  const [user, setUser] = useState(() => {
    try {
      const raw = localStorage.getItem('star_user')
      return raw ? JSON.parse(raw) : null
    } catch {
      return null
    }
  })

  // 同步 user 到 localStorage
  useEffect(() => {
    if (user) {
      localStorage.setItem('star_user', JSON.stringify(user))
    } else {
      localStorage.removeItem('star_user')
    }
  }, [user])

  // 派生值 (依赖 user). isAdmin 是组件里最常用的判断条件
  const value = useMemo(() => ({
    user,
    isAdmin: user?.role === 'admin',
    // 登录: 拉 token 存 localStorage, setUser 触发组件重渲染
    async login(phone, code) {
      const r = await api.login(phone, code)
      const u = r.data.user
      const t = r.data.token
      localStorage.setItem('star_token', t)
      setUser(u)
      return u
    },
    // 退出: 清 token + user
    logout() {
      localStorage.removeItem('star_token')
      setUser(null)
    }
  }), [user])

  return <UserCtx.Provider value={value}>{children}</UserCtx.Provider>
}

// 简单 useContext 包装, 防止在 Provider 之外使用
export function useUser() {
  const ctx = useContext(UserCtx)
  if (!ctx) throw new Error('useUser must be used within UserProvider')
  return ctx
}