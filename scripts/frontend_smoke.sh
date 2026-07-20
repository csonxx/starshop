#!/usr/bin/env bash
echo "=== 前端关键资源加载测试 ==="
for path in \
  "/src/App.jsx" \
  "/src/pages/Home.jsx" \
  "/src/pages/CaseDetail.jsx" \
  "/src/pages/Login.jsx" \
  "/src/pages/Admin/Overview.jsx" \
  "/src/pages/Admin/Banners.jsx" \
  "/src/pages/Admin/Tags.jsx" \
  "/src/pages/Admin/Cases.jsx" \
  "/src/pages/Admin/CaseEditor.jsx" \
  "/src/components/Banner.jsx" \
  "/src/components/CaseCard.jsx" \
  "/src/components/Header.jsx" \
  "/src/styles/global.css" \
  "/src/store/user.jsx" \
  "/src/api/index.js" \
  "/src/api/http.js" \
  "/src/main.jsx" ; do
  code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:5173$path")
  echo "  $code  $path"
done