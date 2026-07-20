// App.jsx - 根组件, 声明所有路由表
//
// 路由结构:
//   /                  首页
//   /cases             案例库 (全部, 不指定风格)
//   /style/:slug       一级风格专题页 (如 /style/cream 奶油风)
//   /cases/:id         案例详情 (任意字符串 ID)
//   /login             登录页
//   /me                用户中心
//   /admin/*           后台路由 (嵌套路由, AdminLayout 提供 chrome)
//     ├ /              概览
//     ├ /banners       Banner CRUD
//     ├ /tags          标签 CRUD
//     ├ /cases         案例列表
//     ├ /cases/new     新建案例
//     └ /cases/:id     编辑案例
import { Route, Routes } from 'react-router-dom'
import Header from './components/Header'
import Home from './pages/Home'
import StylePage from './pages/StylePage'
import CaseDetail from './pages/CaseDetail'
import Login from './pages/Login'
import Me from './pages/Me'
import AdminLayout from './pages/Admin/AdminLayout'
import Overview from './pages/Admin/Overview'
import Banners from './pages/Admin/Banners'
import Tags from './pages/Admin/Tags'
import Cases from './pages/Admin/Cases'
import CaseEditor from './pages/Admin/CaseEditor'

export default function App() {
  return (
    <div className="app-shell">
      {/* 顶层: 后台独立, 其余都走 PublicRoutes */}
      <Routes>
        <Route path="/admin/*" element={<AdminRoutes />} />
        <Route path="*" element={<PublicRoutes />} />
      </Routes>
    </div>
  )
}

// 公开路由: 带 Header
function PublicRoutes() {
  return (
    <>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/cases" element={<StylePage />} />
        <Route path="/style/:slug" element={<StylePage />} />
        <Route path="/cases/:id" element={<CaseDetail />} />
        <Route path="/login" element={<Login />} />
        <Route path="/me" element={<Me />} />
        <Route path="*" element={<Home />} />
      </Routes>
    </>
  )
}

// 后台路由: AdminLayout 嵌套 (左侧菜单 + 内容区)
function AdminRoutes() {
  return (
    <Routes>
      <Route element={<AdminLayout />}>
        <Route index element={<Overview />} />
        <Route path="banners" element={<Banners />} />
        <Route path="tags" element={<Tags />} />
        <Route path="cases" element={<Cases />} />
        <Route path="cases/new" element={<CaseEditor />} />
        <Route path="cases/:id" element={<CaseEditor />} />
      </Route>
    </Routes>
  )
}