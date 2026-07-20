import { useEffect, useState } from 'react'
import { api } from '../../api'

export default function Overview() {
  const [stats, setStats] = useState(null)
  const [byStyle, setByStyle] = useState([])

  useEffect(() => {
    api.adminOverview().then((r) => setStats(r.data)).catch(console.error)
    api.adminStatsByStyle().then((r) => setByStyle(r.data || [])).catch(console.error)
  }, [])

  return (
    <div>
      <h1 className="adm-h1">数据概览</h1>
      <p className="adm-sub">星仔高端定制 · 运营后台实时统计</p>

      <div className="adm-stat-grid">
        <Stat n={stats?.bannerCount} label="Banner 数量" />
        <Stat n={stats?.styleTagCount} label="一级风格标签" />
        <Stat n={stats?.tagCount} label="全部标签" />
        <Stat n={stats?.caseCount} label="案例总数" />
        <Stat n={stats?.pinnedCount} label="置顶案例" />
      </div>

      <div className="adm-card">
        <div className="adm-card-h">
          <h2>各风格案例分布</h2>
        </div>
        <table className="adm-table">
          <thead>
            <tr>
              <th>风格</th>
              <th>案例数</th>
              <th>占比</th>
              <th>进度</th>
            </tr>
          </thead>
          <tbody>
            {byStyle.map((s) => {
              const total = byStyle.reduce((a, b) => a + b.value, 0) || 1
              const pct = (s.value / total) * 100
              return (
                <tr key={s.name}>
                  <td><strong>{s.name}</strong></td>
                  <td>{s.count}</td>
                  <td>{pct.toFixed(1)}%</td>
                  <td>
                    <div style={{
                      width: 200,
                      height: 6,
                      background: 'var(--mist)',
                      borderRadius: 3,
                      overflow: 'hidden'
                    }}>
                      <div style={{
                        width: `${Math.max(pct, 2)}%`,
                        height: '100%',
                        background: 'linear-gradient(90deg, var(--gold) 0%, var(--champagne) 100%)'
                      }} />
                    </div>
                  </td>
                </tr>
              )
            })}
            {byStyle.length === 0 && (
              <tr><td colSpan={4} style={{ textAlign: 'center', padding: 40 }}>暂无数据</td></tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function Stat({ n, label }) {
  return (
    <div className="adm-stat">
      <div className="adm-stat-label">{label}</div>
      <div className="adm-stat-n">{n ?? '-'}<em>条</em></div>
    </div>
  )
}