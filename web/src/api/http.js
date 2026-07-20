import axios from 'axios'

export const http = axios.create({
  baseURL: '/api/v1',
  timeout: 15000
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('star_token')
  if (token) {
    config.headers = config.headers || {}
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (resp) => {
    const data = resp.data
    if (data && typeof data === 'object' && 'code' in data) {
      if (data.code === 0) return data
      return Promise.reject(new Error(data.msg || 'request failed'))
    }
    return data
  },
  (err) => {
    const status = err.response?.status
    const msg = err.response?.data?.msg || err.message
    if (status === 401) {
      localStorage.removeItem('star_token')
      localStorage.removeItem('star_user')
      if (!location.pathname.startsWith('/login')) {
        location.href = '/login?redirect=' + encodeURIComponent(location.pathname)
      }
    }
    return Promise.reject(new Error(msg || '网络异常'))
  }
)