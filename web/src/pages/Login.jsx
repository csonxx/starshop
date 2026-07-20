import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { api } from '../api'
import { useUser } from '../store/user.jsx'
import './Login.css'

export default function Login() {
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [hint, setHint] = useState('')
  const [loading, setLoading] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const nav = useNavigate()
  const [params] = useSearchParams()
  const redirect = params.get('redirect') || '/'
  const { login } = useUser()

  const sendCode = async () => {
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      setHint('请输入有效的手机号')
      return
    }
    try {
      const r = await api.sendCode(phone)
      setHint(`验证码已发送 (开发期固定 ${r.data.code})`)
      setCountdown(60)
      const timer = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) { clearInterval(timer); return 0 }
          return c - 1
        })
      }, 1000)
    } catch (e) {
      setHint(e.message)
    }
  }

  const onSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      const u = await login(phone, code)
      if (u.role === 'admin') {
        nav('/admin', { replace: true })
      } else {
        nav(redirect, { replace: true })
      }
    } catch (err) {
      setHint(err.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="lg">
      <div className="lg-card">
        <div className="lg-l">
          <div className="lg-brand">
            <span className="brand-mark">星</span>
            <span>
              <strong className="display">星仔高端定制</strong>
              <small>工厂直营 · 全屋定制</small>
            </span>
          </div>
          <h1 className="lg-title">欢迎来到<br/>星仔高端定制</h1>
          <p className="lg-desc">为热爱生活的您，定制一座理想之家。</p>
          <div className="lg-features">
            <div><span>·</span> 12,000㎡ 自有智造工厂</div>
            <div><span>·</span> 设计师 1V1 全程跟进</div>
            <div><span>·</span> 平均 28 天交付到家</div>
            <div><span>·</span> 10 年质保终身维护</div>
          </div>
        </div>
        <div className="lg-r">
          <h2 className="serif">手机号登录</h2>
          <p className="lg-tip">输入手机号获取验证码 · 管理员账号 13800138000</p>
          <form className="lg-form" onSubmit={onSubmit}>
            <div className="lg-field">
              <label>手机号</label>
              <input
                type="tel"
                inputMode="numeric"
                maxLength={11}
                value={phone}
                onChange={(e) => setPhone(e.target.value.replace(/\D/g, ''))}
                placeholder="请输入手机号"
                autoFocus
              />
            </div>
            <div className="lg-field">
              <label>验证码</label>
              <div className="lg-code-row">
                <input
                  type="text"
                  inputMode="numeric"
                  maxLength={6}
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, ''))}
                  placeholder="6 位验证码"
                />
                <button
                  type="button"
                  className="lg-send"
                  onClick={sendCode}
                  disabled={countdown > 0}
                >
                  {countdown > 0 ? `${countdown}s 后重试` : '获取验证码'}
                </button>
              </div>
            </div>

            {hint && <div className="lg-hint">{hint}</div>}

            <button
              type="submit"
              className="btn btn-gold btn-block"
              disabled={loading || !phone || !code}
            >
              {loading ? '登录中...' : '登 录'}
            </button>

            <div className="lg-foot">
              登录即同意 <a href="#">《用户协议》</a> 与 <a href="#">《隐私政策》</a>
            </div>
          </form>

          <div className="lg-debug">
            <strong>开发提示</strong>
            验证码固定为 <code>1234</code> ；不同手机号对应不同角色：
            <ul>
              <li><code>13800138000</code> → 管理员 · 进入运营后台</li>
              <li><code>13900000001</code> / <code>13900000002</code> → 销售 · 看精准价</li>
              <li><code>13700000001</code> / <code>13700000002</code> → 供应商 · 看精准价</li>
              <li>其他手机号 → 普通用户 · 仅看价格区间</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  )
}