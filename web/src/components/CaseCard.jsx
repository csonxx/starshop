import { Link } from 'react-router-dom'
import { useUser } from '../store/user.jsx'
import './CaseCard.css'

// 渲染价格: 普通用户 (price=0) 显示价格区间; 销售/供应商/管理员显示精准数字
// .num 类专门处理数字字体, 防止装饰性 display/serif 字体的 1 与 I 混淆
function renderPrice(item) {
  if (!item.price || item.price === 0) {
    return <span className="price-num">{item.priceLabel || '请询价'}</span>
  }
  return (
    <span className="price-num">
      <span className="num-sym">¥</span><span className="num">{(item.price / 10000).toFixed(2)}</span><em>万</em>
    </span>
  )
}

export default function CaseCard({ item }) {
  const { isAdmin } = useUser()
  return (
    <Link to={`/cases/${item.id}`} className="case-card fade-up">
      <div className="case-cover">
        <img src={item.cover} alt={item.title} loading="lazy" />
        {item.pinned && <span className="case-pinned">🔥 爆款</span>}
        <div className="case-cover-shade" />
        <div className="case-quick">
          <span>{item.style}</span>
          <span>·</span>
          <span>{item.space}</span>
          <span>·</span>
          <span>{item.size}</span>
        </div>
      </div>
      <div className="case-meta">
        <h3 className="case-title">{item.title}</h3>
        <div className="case-row">
          <div className="case-colors">
            {(item.colors || []).slice(0, 4).map((c) => (
              <span key={c} className="mini-dot" data-color={c} />
            ))}
          </div>
          <div className="case-price">
            <span className="price-from">{item.price ? '精准价' : '参考价'}</span>
            {renderPrice(item)}
          </div>
        </div>
        {isAdmin && item.pinned && <div className="admin-hint">🔥 已置顶</div>}
      </div>
    </Link>
  )
}