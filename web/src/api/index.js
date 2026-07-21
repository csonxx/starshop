import { http } from './http'

export const api = {
  sendCode: (phone) => http.post('/auth/send-code', { phone }),
  login: (phone, code) => http.post('/auth/login', { phone, code }),
  me: () => http.get('/me'),

  banners: (config = {}) => http.get('/banners', config),
  tags: (type, config = {}) => http.get('/tags', { ...config, params: type ? { type } : {} }),
  cases: (params, config = {}) => http.get('/cases', { ...config, params }),
  caseDetail: (id, config = {}) => http.get(`/cases/${id}`, config),
  pinned: (config = {}) => http.get('/cases/pinned', config),

  // admin
  adminOverview: () => http.get('/admin/overview'),
  adminStatsByStyle: () => http.get('/admin/stats/by-style'),

  adminListBanners: () => http.get('/admin/banners'),
  adminCreateBanner: (payload) => http.post('/admin/banners', payload),
  adminUpdateBanner: (id, payload) => http.put(`/admin/banners/${id}`, payload),
  adminDeleteBanner: (id) => http.delete(`/admin/banners/${id}`),

  adminListTags: (type) => http.get('/admin/tags', { params: type ? { type } : {} }),
  adminCreateTag: (payload) => http.post('/admin/tags', payload),
  adminUpdateTag: (id, payload) => http.put(`/admin/tags/${id}`, payload),
  adminDeleteTag: (id) => http.delete(`/admin/tags/${id}`),

  adminListCases: () => http.get('/admin/cases'),
  adminGetCase: (id) => http.get(`/admin/cases/${id}`),
  adminCreateCase: (payload) => http.post('/admin/cases', payload),
  adminUpdateCase: (id, payload) => http.put(`/admin/cases/${id}`, payload),
  adminDeleteCase: (id) => http.delete(`/admin/cases/${id}`),
  adminTogglePin: (id) => http.post(`/admin/cases/${id}/pin`)
}
